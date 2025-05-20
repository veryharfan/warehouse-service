package pkg

import "database/sql"

func WithTransaction(tx *sql.Tx, fn func() error) error {
	err := fn()
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
