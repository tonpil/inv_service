package domain

type Service interface {
	CreateDocument(doc *Doc) (*Doc, error)
	GetDocument(id string) (*Doc, error)
	ListDocuments(*int32, *int32) ([]*Doc, error)
	UpdateDocument(doc *Doc, parentID *string) (*Doc, error)
	DeleteDocument(id string) error
}
