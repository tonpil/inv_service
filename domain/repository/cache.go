package repository

import entity "doc_service/domain"

type CacheRepository interface {
	Set(doc *entity.Doc) error
	Get(id string) (*entity.Doc, error)
	Flush()
}
