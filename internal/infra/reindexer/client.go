package reindexer

import (
	"fmt"

	"github.com/restream/reindexer/v5"
	_ "github.com/restream/reindexer/v5/bindings/cproto"
)

func NewClient(host string, port int32, name string, user string, password string) (*reindexer.Reindexer, error) {
	dsn := fmt.Sprintf("cproto://%s:%s@%s:%d/%s", user, password, host, port, name)
	db, err := reindexer.NewReindex(dsn, reindexer.WithCreateDBIfMissing())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
