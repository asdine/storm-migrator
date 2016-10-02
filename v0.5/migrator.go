package storm

import (
	"reflect"

	"github.com/asdine/storm-migrator/v0.5/codec"
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
func (m *Migrator) Run(instances []interface{}) error {
	db, err := Open("", UseDB(m.boltDB))
	if err != nil {
		return err
	}

	for _, inst := range instances {
		// extract informations
		ref := reflect.Indirect(reflect.ValueOf(inst))
		info, err := extract(&ref)
		if err != nil {
			return err
		}

		// create an empty slice
		sliceType := reflect.SliceOf(ref.Type())
		newSlice := reflect.New(sliceType)

		// fetch all the records
		err = db.All(newSlice.Interface())
		if err != nil {
			return err
		}

		// drop the bucket
		err = db.Drop(info.Name)
		if err != nil {
			return err
		}

		// resave all the records
		s := newSlice.Elem()
		l := s.Len()
		for i := 0; i < l; i++ {
			v := s.Index(i)
			err = db.Save(v.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}

	// drop the old metadata bucket
	err = db.Drop(metadataBucket)
	if err != nil {
		return err
	}

	// set new version
	return db.Set(dbinfo, "version", "0.5.0")
}
