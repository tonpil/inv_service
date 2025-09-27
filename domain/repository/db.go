package repository

import (
	entity "doc_service/domain"
)

type DBRepository interface {
	List(limit int32, offset int32) ([]*entity.Doc, error)
	ListByIDs(IDs []string) ([]*entity.Doc, error)
	ListByChildrenIDs(childrenIDs []string) ([]*entity.Doc, error)
	GetByID(id string) (*entity.Doc, error)
	GetByChildrenID(childrenID string) (*entity.Doc, error)
	Create(doc *entity.Doc) error
	Update(tx any, doc *entity.Doc) error
	Delete(tx any, id string) error
}

type UnitOfWork interface {
	Begin() (any, error)
	Commit(tx any) error
	RollBack(tx any) error
}
