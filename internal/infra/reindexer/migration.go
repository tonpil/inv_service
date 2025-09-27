package reindexer

import (
	entity "doc_service/domain"

	"github.com/restream/reindexer/v5"
)

const (
	docsNamespace = "docs" // TODO: remove hardcode
)

func InitNamespaces(db *reindexer.Reindexer) error {
	err := db.OpenNamespace(
		docsNamespace,
		reindexer.DefaultNamespaceOptions(),
		entity.Doc{},
	)
	if err != nil {
		return err
	}

	return nil
}
