package migrator

import (
	"fmt"
	"io"
	"os"
	"time"

	stormv05 "github.com/asdine/storm-migrator/v0.5"
	"github.com/asdine/storm-migrator/v0.5/codec"
	"github.com/boltdb/bolt"
)

// New instanciates a Migrator for the given database
func New(path string) *Migrator {
	return &Migrator{
		path: path,
	}
}

// Migrator handles database migration for databases that use old versions of Storm
type Migrator struct {
	path       string
	instances  []interface{}
	forceCodec codec.MarshalUnmarshaler
}

// AddBucket registers a bucket to migrate based on the given instance.
func (m *Migrator) AddBucket(instance interface{}) {
	m.instances = append(m.instances, instance)
}

//
// // AddBuckets registers buckets to migrate based on the given instances.
// // Must be used for the buckets created with Save or Init.
// func (m *Migrator) AddBuckets(instances ...interface{}) {
// 	m.instances = append(m.instances, instances...)
// }

//
// // AddKV registers a key value pair to migrate based on the bucket name, the inst.
// // Must be used for the buckets created using Set.
// func (m *Migrator) AddKV(bucketName string, keyInstance interface{}, valueInstances ...interface{}) {
// 	m.instances = instances
// }

// Run the migration
func (m *Migrator) Run(dst string, options ...func(*Migrator) error) error {
	for _, option := range options {
		err := option(m)
		if err != nil {
			return err
		}
	}

	_, err := os.Stat(dst)
	if err == nil {
		return fmt.Errorf("Path \"%s\" already exists.", dst)
	}

	err = m.checkSourceDB()
	if err != nil {
		return err
	}

	err = m.copyDB(dst)
	if err != nil {
		return err
	}

	b, err := bolt.Open(dst, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	migrator := stormv05.NewMigrator(b, m.forceCodec)
	return migrator.Run(m.instances)
}

func (m *Migrator) checkSourceDB() error {
	_, err := os.Stat(m.path)
	if err != nil {
		return err
	}

	db, err := bolt.Open(m.path, 0600, &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true})
	if err != nil {
		return err
	}

	return db.Close()
}

func (m *Migrator) copyDB(path string) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}

	src, err := os.Open(m.path)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return dst.Sync()
}

// Codec option forces the codec used for the whole migration
func Codec(codec codec.MarshalUnmarshaler) func(*Migrator) error {
	return func(m *Migrator) error {
		m.forceCodec = codec
		return nil
	}
}
