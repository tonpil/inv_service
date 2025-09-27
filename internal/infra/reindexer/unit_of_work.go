package reindexer

import "github.com/restream/reindexer/v5"

type UnitOfWork struct {
	db *reindexer.Reindexer
	ns string
}

func NewUnitOfWork(db *reindexer.Reindexer) *UnitOfWork {
	return &UnitOfWork{
		db: db,
		ns: docsNamespace, // TODO: remove hardcode
	}
}

func (u *UnitOfWork) Begin() (any, error) {
	tx, err := u.db.BeginTx(u.ns)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (u *UnitOfWork) Commit(tx any) error {
	return tx.(*reindexer.Tx).Commit()
}

func (u *UnitOfWork) RollBack(tx any) error {
	return tx.(*reindexer.Tx).Rollback()
}
