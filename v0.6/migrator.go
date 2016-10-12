package storm

import (
	"github.com/asdine/storm-migrator/v0.6/codec"
	"github.com/boltdb/bolt"
)

// NewMigrator instantiates a new Migrator
func NewMigrator(db *bolt.DB, codec codec.MarshalUnmarshaler) *Migrator {
	return &Migrator{boltDB: db, codec: codec}
}

// Migrator migrates the given database to v0.5
type Migrator struct {
	boltDB *bolt.DB
	codec  codec.MarshalUnmarshaler
}

// Run the migration
func (m *Migrator) Run(instances []interface{}, kvKeys map[string][]interface{}) error {
	db, err := Open("", UseDB(m.boltDB))
	if err != nil {
		return err
	}

	err = m.runSaved(db, instances)
	if err != nil {
		return err
	}

	// set new version
	return db.Set(dbinfo, "version", Version)
}

func (m *Migrator) runSaved(db *DB, instances []interface{}) error {
	for _, inst := range instances {
		// reindex
		err := db.ReIndex(inst)
		if err != nil {
			return err
		}
	}

	return nil
}
