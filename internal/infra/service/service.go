package service

import (
	"doc_service/domain"
	"doc_service/domain/repository"
	"errors"
	"sync"
)

const (
	defaultLimit  = 10
	defaultOffset = 0
)

type DocumentService struct {
	DBRepository    repository.DBRepository
	UnitOfWork      repository.UnitOfWork
	CacheRepository repository.CacheRepository
}

func NewDocumentService(dbRepository repository.DBRepository, unitOfWork repository.UnitOfWork, cacheRepository repository.CacheRepository) *DocumentService {
	return &DocumentService{
		DBRepository:    dbRepository,
		UnitOfWork:      unitOfWork,
		CacheRepository: cacheRepository,
	}
}

func (s *DocumentService) CreateDocument(doc *domain.Doc) (*domain.Doc, error) {
	if doc.ID == "" || doc.Title == "" || doc.Sort <= 0 {
		return nil, domain.ErrInvalidArguments
	}

	var (
		subDocuments, parentDocuments                    []*domain.Doc
		document                                         *domain.Doc
		subDocumentsErr, parentDocumentsErr, documentErr error
		wg                                               sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		document, documentErr = s.DBRepository.GetByID(doc.ID)
	}()

	go func() {
		defer wg.Done()
		subDocuments, subDocumentsErr = s.DBRepository.ListByIDs(doc.ChildrenIDs)
	}()

	go func() {
		defer wg.Done()
		parentDocuments, parentDocumentsErr = s.DBRepository.ListByChildrenIDs(doc.ChildrenIDs)
	}()

	wg.Wait()

	if subDocumentsErr != nil {
		return nil, subDocumentsErr
	}

	if parentDocumentsErr != nil {
		return nil, parentDocumentsErr
	}

	if documentErr != nil {
		return nil, documentErr
	}

	if len(parentDocuments) > 0 {
		return nil, domain.ErrInvalidArguments
	}

	if len(subDocuments) != len(doc.ChildrenIDs) {
		return nil, domain.ErrNotFound
	}

	if document != nil {
		return nil, domain.ErrAlreadyExists
	}

	for _, subDocument := range subDocuments {
		if len(subDocument.ChildrenIDs) > 0 {
			subSubDocuments, err := s.DBRepository.ListByIDs(subDocument.ChildrenIDs)
			if err != nil {
				return nil, err
			}

			for _, subSubDocument := range subSubDocuments {
				if len(subSubDocument.ChildrenIDs) > 0 {
					return nil, domain.ErrInvalidArguments
				}
			}
		}
	}

	err := s.DBRepository.Create(doc)
	if err != nil {
		return nil, err
	}

	doc.SubDocs = subDocuments
	doc.Finalize()

	s.CacheRepository.Flush()

	return doc, nil
}

func (s *DocumentService) ListDocuments(limit *int32, offset *int32) ([]*domain.Doc, error) {
	var limitToDB, offsetToDB int32

	if limit == nil {
		limitToDB = defaultLimit
	} else {
		limitToDB = *limit
	}

	if offset == nil {
		offsetToDB = defaultOffset
	} else {
		offsetToDB = *offset
	}

	rootDocs, err := s.DBRepository.List(limitToDB, offsetToDB)
	if err != nil {
		return nil, err
	}
	if len(rootDocs) == 0 {
		return []*domain.Doc{}, nil
	}

	for _, rootDoc := range rootDocs {
		subDocumentIDs := make([]string, 0, len(rootDoc.ChildrenIDs))
		for _, subDoc := range rootDoc.SubDocs {
			subDocumentIDs = append(subDocumentIDs, subDoc.ChildrenIDs...)
		}

		subDocuments, err := s.DBRepository.ListByIDs(subDocumentIDs)
		if err != nil {
			return nil, err
		}

		subDocumentsMap := make(map[string]*domain.Doc)
		for _, subDoc := range subDocuments {
			subDocumentsMap[subDoc.ID] = subDoc
		}

		for _, subDoc := range rootDoc.SubDocs {
			subDoc.SubDocs = make([]*domain.Doc, 0, len(subDoc.ChildrenIDs))
			for _, childrenID := range subDoc.ChildrenIDs {
				if child, ok := subDocumentsMap[childrenID]; ok {
					subDoc.SubDocs = append(subDoc.SubDocs, child)
				}
			}
		}
	}

	var wg sync.WaitGroup

	wg.Add(len(rootDocs))

	for i := range rootDocs {
		root := rootDocs[i]

		go func() {

			defer wg.Done()
			root.Finalize()
		}()
	}

	wg.Wait()

	return rootDocs, nil
}

func (s *DocumentService) GetDocument(id string) (*domain.Doc, error) {
	var rootDocument *domain.Doc

	rootDocument, err := s.CacheRepository.Get(id)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
	} else {
		return rootDocument, nil
	}

	rootDocument, err = s.DBRepository.GetByID(id)
	if err != nil {
		return nil, err
	}

	if rootDocument == nil {
		return nil, domain.ErrNotFound
	}

	subDocumentIDs := make([]string, 0, len(rootDocument.ChildrenIDs))
	for _, subDoc := range rootDocument.SubDocs {
		subDocumentIDs = append(subDocumentIDs, subDoc.ChildrenIDs...)
	}

	subDocuments, err := s.DBRepository.ListByIDs(subDocumentIDs)
	if err != nil {
		return nil, err
	}

	subDocumentsMap := make(map[string]*domain.Doc)
	for _, subDoc := range subDocuments {
		subDocumentsMap[subDoc.ID] = subDoc
	}

	for _, subDoc := range rootDocument.SubDocs {
		subDoc.SubDocs = make([]*domain.Doc, 0, len(subDoc.ChildrenIDs))
		for _, childrenID := range subDoc.ChildrenIDs {
			if child, ok := subDocumentsMap[childrenID]; ok {
				subDoc.SubDocs = append(subDoc.SubDocs, child)
			}
		}
	}

	rootDocument.Finalize()

	err = s.CacheRepository.Set(rootDocument)
	if err != nil {
		return nil, err
	}

	return rootDocument, nil
}

func (s *DocumentService) UpdateDocument(doc *domain.Doc, parentID *string) (*domain.Doc, error) {
	if doc.Sort <= 0 {
		return nil, domain.ErrInvalidArguments
	}

	oldDoc, err := s.DBRepository.GetByID(doc.ID)
	if err != nil {
		return nil, err
	}
	if oldDoc == nil {
		return nil, domain.ErrNotFound
	}

	currentParent, err := s.DBRepository.GetByChildrenID(oldDoc.ID)
	if err != nil {
		return nil, err
	}

	var newParent *domain.Doc
	if parentID != nil {
		if *parentID == doc.ID {
			return nil, domain.ErrInvalidArguments
		}
		newParent, err = s.DBRepository.GetByID(*parentID)
		if err != nil {
			return nil, err
		}
		if newParent == nil {
			return nil, domain.ErrNotFound
		}
	}

	var removed []string
	var added []string
	if currentParent != nil && len(oldDoc.ChildrenIDs) > 0 {
		oldIDs := make(map[string]struct{})
		newIDs := make(map[string]struct{})

		for _, sub := range oldDoc.SubDocs {
			oldIDs[sub.ID] = struct{}{}
		}
		for _, sub := range doc.SubDocs {
			newIDs[sub.ID] = struct{}{}
		}

		removed = make([]string, 0, len(oldIDs))
		added = make([]string, 0, len(newIDs))
		for _, childrenID := range oldDoc.ChildrenIDs {
			if _, ok := newIDs[childrenID]; !ok {
				removed = append(removed, childrenID)
			}
		}

		for _, childrenID := range doc.ChildrenIDs {
			if _, ok := oldIDs[childrenID]; !ok {
				added = append(added, childrenID)
			}
		}
	} else {
		added = doc.ChildrenIDs
	}

	notAvalibleSubDocs, err := s.DBRepository.ListByChildrenIDs(added)
	if err != nil {
		return nil, err
	}
	if len(notAvalibleSubDocs) > 0 {
		return nil, domain.ErrInvalidArguments
	}

	newSubDocs, err := s.DBRepository.ListByIDs(doc.ChildrenIDs)
	if err != nil {
		return nil, err
	}

	for _, subDocument := range newSubDocs {
		if len(subDocument.ChildrenIDs) > 0 {
			subSubDocuments, err := s.DBRepository.ListByIDs(subDocument.ChildrenIDs)
			if err != nil {
				return nil, err
			}

			subDocument.SubDocs = subSubDocuments

			for _, subSubDocument := range subSubDocuments {
				if len(subSubDocument.ChildrenIDs) > 0 {
					return nil, domain.ErrInvalidArguments
				}
			}
		}
	}

	doc.SubDocs = newSubDocs
	if newParent != nil {
		err = validateAfterAttach(doc, newParent, s.DBRepository)
		if err != nil {
			return nil, err
		}
	}

	tx, err := s.UnitOfWork.Begin()
	if err != nil {
		return nil, err
	}

	if currentParent != nil {
		childrenIDs := make([]string, 0, len(currentParent.ChildrenIDs)-1+len(removed))
		childrenIDs = append(childrenIDs, currentParent.ChildrenIDs...)
		childrenIDs = append(childrenIDs, removed...)
		currentParent.ChildrenIDs = removeString(childrenIDs, doc.ID)
		err = s.DBRepository.Update(tx, currentParent)
		if err != nil {
			s.UnitOfWork.RollBack(tx)
			return nil, err
		}
	}

	if newParent != nil {
		childrenIDs := make([]string, 0, len(newParent.ChildrenIDs)+1)
		childrenIDs = append(childrenIDs, newParent.ChildrenIDs...)
		childrenIDs = append(childrenIDs, doc.ID)
		newParent.ChildrenIDs = childrenIDs

		err = s.DBRepository.Update(tx, newParent)
		if err != nil {
			s.UnitOfWork.RollBack(tx)
			return nil, err
		}

		if currentParent != nil {
			currentParent.ChildrenIDs = removeString(currentParent.ChildrenIDs, doc.ID)
			err = s.DBRepository.Update(tx, currentParent)
			if err != nil {
				s.UnitOfWork.RollBack(tx)
				return nil, err
			}
		}

	}

	err = s.DBRepository.Update(tx, doc)
	if err != nil {
		s.UnitOfWork.RollBack(tx)
		return nil, err
	}
	err = s.UnitOfWork.Commit(tx)
	if err != nil {
		s.UnitOfWork.RollBack(tx)
		return nil, err
	}

	s.CacheRepository.Flush()

	doc.Finalize()
	return doc, nil
}

func (s *DocumentService) DeleteDocument(id string) error {
	var (
		docToDelete, parentDoc       *domain.Doc
		docToDeleteErr, parentDocErr error
		wg                           sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		docToDelete, docToDeleteErr = s.DBRepository.GetByID(id)
	}()

	go func() {
		defer wg.Done()
		parentDoc, parentDocErr = s.DBRepository.GetByChildrenID(id)
	}()

	wg.Wait()

	if docToDeleteErr != nil {
		return docToDeleteErr
	}

	if parentDocErr != nil {
		return parentDocErr
	}

	if docToDelete == nil {
		return domain.ErrNotFound
	}

	tx, err := s.UnitOfWork.Begin()
	if err != nil {
		return err
	}

	if parentDoc != nil {
		parentDoc.ChildrenIDs = removeString(parentDoc.ChildrenIDs, id)
		parentDoc.ChildrenIDs = append(parentDoc.ChildrenIDs, docToDelete.ChildrenIDs...)
		if err := s.DBRepository.Update(tx, parentDoc); err != nil {
			s.UnitOfWork.RollBack(tx)
			return err
		}

	}

	err = s.DBRepository.Delete(tx, id)
	if err != nil {
		s.UnitOfWork.RollBack(tx)
		return err
	}

	err = s.UnitOfWork.Commit(tx)
	if err != nil {
		s.UnitOfWork.RollBack(tx)
		return err
	}

	s.CacheRepository.Flush() // TODO: add invalidation?

	return nil
}

func removeString(slice []string, target string) []string {
	for i, v := range slice {
		if v == target {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func validateAfterAttach(doc *domain.Doc, newParent *domain.Doc, repo repository.DBRepository) error {
	root := newParent

	root.SubDocs = append(root.SubDocs, doc)

	for {
		p, err := repo.GetByChildrenID(root.ID)
		if err != nil {
			return err
		}
		if p == nil {
			break
		}

		p.SubDocs = append(p.SubDocs, root)
		root = p
	}

	if err := validateHierarchy(root, 0); err != nil {
		return err
	}

	return nil
}

func validateHierarchy(doc *domain.Doc, level int) error {
	if level > 2 {
		return domain.ErrInvalidArguments
	}
	for _, sub := range doc.SubDocs {
		if err := validateHierarchy(sub, level+1); err != nil {
			return err
		}
	}
	return nil
}
