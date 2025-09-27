package reindexer

import (
	"fmt"

	entity "doc_service/domain"

	"github.com/restream/reindexer/v5"
)

type Repository struct {
	db *reindexer.Reindexer
	ns string
}

func NewRepository(db *reindexer.Reindexer) *Repository {
	return &Repository{
		db: db,
		ns: docsNamespace, // TODO: remove hardcode
	}
}

func (r *Repository) List(limit int32, offset int32) ([]*entity.Doc, error) {
	it := r.db.Query(r.ns).
		Not().
		Where(
			"id",
			reindexer.SET,
			r.db.Query(r.ns).Select("children_ids"),
		).
		Offset(int(offset)).
		Limit(int(limit)).
		LeftJoin(r.db.Query(r.ns), "sub_docs").
		On("children_ids", reindexer.SET, "id").
		Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	var docs []*entity.Doc
	for it.Next() {
		doc := it.Object().(*entity.Doc)
		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *Repository) ListByIDs(IDs []string) ([]*entity.Doc, error) {
	it := r.db.Query(r.ns).
		Where("id", reindexer.SET, IDs).
		LeftJoin(r.db.Query(r.ns), "sub_docs").
		On("children_ids", reindexer.SET, "id").
		Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	var docs []*entity.Doc
	for it.Next() {
		doc := it.Object().(*entity.Doc)
		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *Repository) ListByChildrenIDs(childrenIDs []string) ([]*entity.Doc, error) {
	it := r.db.Query(r.ns).
		Where("children_ids", reindexer.SET, childrenIDs).
		LeftJoin(r.db.Query(r.ns), "sub_docs").
		On("children_ids", reindexer.SET, "id").
		Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	var docs []*entity.Doc
	for it.Next() {
		doc := it.Object().(*entity.Doc)
		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *Repository) GetByID(id string) (*entity.Doc, error) {
	it := r.db.Query(r.ns).
		Where("id", reindexer.EQ, id).
		LeftJoin(r.db.Query(r.ns), "sub_docs").
		On("children_ids", reindexer.SET, "id").
		Exec()
	defer it.Close()

	err := it.Error()

	if err != nil {
		return nil, err
	}

	if !it.Next() {
		return nil, nil
	}

	return it.Object().(*entity.Doc), nil
}

func (r *Repository) GetByChildrenID(childrenID string) (*entity.Doc, error) {
	it := r.db.Query(r.ns).
		Where("children_ids", reindexer.SET, childrenID).
		LeftJoin(r.db.Query(r.ns), "sub_docs").
		On("children_ids", reindexer.SET, "id").
		Exec()
	defer it.Close()

	err := it.Error()

	if err != nil {
		return nil, err
	}

	if !it.Next() {
		return nil, nil
	}

	return it.Object().(*entity.Doc), nil
}

func (r *Repository) Create(doc *entity.Doc) error {
	if _, err := r.db.Insert(r.ns, doc); err != nil {
		return fmt.Errorf("failed to create doc: %w", err)
	}
	return nil
}

func (r *Repository) Update(tx any, doc *entity.Doc) error {
	if err := tx.(*reindexer.Tx).Upsert(doc); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Delete(tx any, id string) error {
	if err := tx.(*reindexer.Tx).Delete(entity.Doc{ID: id}); err != nil {
		return err
	}
	return nil
}
