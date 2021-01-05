package boltdb

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/teseraio/ensemble/operator/state"
)

// Factory is the factory method for the Boltdb backend
func Factory(config map[string]interface{}) (state.State, error) {
	pathRaw, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("field 'path' not found")
	}
	path, ok := pathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("field 'path' is not string")
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	b := &BoltDB{
		db: db,
	}
	return b, nil
}

// BoltDB is a boltdb state implementation
type BoltDB struct {
	db *bolt.DB
}

// Close implements the BoltDB interface
func (b *BoltDB) Close() error {
	return b.db.Close()
}
