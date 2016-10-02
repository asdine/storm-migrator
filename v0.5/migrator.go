package storm

import (
	"reflect"

	"github.com/asdine/storm-migrator/v0.5/codec"
	"github.com/asdine/storm-migrator/v0.5/q"
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

	err = m.runSet(db, kvKeys)
	if err != nil {
		return err
	}

	// drop the old metadata bucket
	err = db.Drop(metadataBucket)
	if err != nil {
		return err
	}

	// set new version
	return db.Set(dbinfo, "version", "0.5.0")
}

func (m *Migrator) runSaved(db *DB, instances []interface{}) error {
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

	return nil
}

func (m *Migrator) runSet(db *DB, kvKeys map[string][]interface{}) error {
	for bucketName, instances := range kvKeys {
		// fetch all keys and values from the bucket
		var keys [][]byte
		var values [][]byte
		err := db.Select(q.True()).Bucket(bucketName).RawEach(func(k []byte, v []byte) error {
			cpk := make([]byte, len(k))
			cpv := make([]byte, len(v))
			copy(cpk, k)
			copy(cpv, v)
			keys = append(keys, cpk)
			values = append(values, cpv)
			return nil
		})
		if err != nil {
			return err
		}

		for i, k := range keys {
			// find the right instance
			for _, inst := range instances {
				r := reflect.Indirect(reflect.ValueOf(inst))
				t := r.Type()
				switch {
				case t.Kind() == reflect.String:
					r.SetString(string(k))
				case t.AssignableTo(reflect.TypeOf([]byte{})):
					r.SetBytes(k)
				default:
					err = m.codec.Unmarshal(k, inst)
					if err != nil {
						// if it doesn't match, try another
						continue
					}
				}

				// create new key
				newKey, err := toBytes(r.Interface(), m.codec)
				if err != nil {
					return err
				}

				tx, err := db.Bolt.Begin(true)
				if err != nil {
					return err
				}

				b := tx.Bucket([]byte(bucketName))

				// delete the old record
				err = b.Delete(k)
				if err != nil {
					tx.Rollback()
					return err
				}

				// save the new record
				err = b.Put(newKey, values[i])
				if err != nil {
					tx.Rollback()
					return err
				}

				tx.Commit()
				break
			}
		}
	}

	return nil
}
