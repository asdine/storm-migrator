package migrator

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	stormv05 "github.com/asdine/storm-migrator/v0.5"
	"github.com/asdine/storm-migrator/v0.5/codec"
	"github.com/asdine/storm-migrator/v0.5/codec/json"
	stormv06 "github.com/asdine/storm-migrator/v0.6"
	"github.com/boltdb/bolt"
)

// New instanciates a Migrator for the given database
func New(path string) *Migrator {
	return &Migrator{
		path:       path,
		kvKeys:     make(map[string][]interface{}),
		forceCodec: json.Codec,
	}
}

// Migrator handles database migration for databases that use old versions of Storm
type Migrator struct {
	path       string
	instances  []interface{}
	kvKeys     map[string][]interface{}
	forceCodec codec.MarshalUnmarshaler
}

// AddBuckets registers buckets to migrate based on the given instances.
// Must be used for the buckets created with Save or Init.
func (m *Migrator) AddBuckets(instances ...interface{}) {
	m.instances = append(m.instances, instances...)
}

// AddKV registers a key value pair to migrate based on the bucket name, the inst.
// Must be used for the buckets created using Set.
func (m *Migrator) AddKV(bucketName string, keyInstances []interface{}) {
	m.kvKeys[bucketName] = append(m.kvKeys[bucketName], keyInstances...)
}

// Run the migration.
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
	defer b.Close()

	for {
		version, err := m.getVersion(b)
		if err != nil {
			fmt.Println("1")
			return err
		}

		switch {
		case strings.HasPrefix(version, "0.4"):
			migrator := stormv05.NewMigrator(b, m.forceCodec)
			err = migrator.Run(m.instances, m.kvKeys)
		case strings.HasPrefix(version, "0.5"):
			migrator := stormv06.NewMigrator(b, m.forceCodec)
			err = migrator.Run(m.instances, m.kvKeys)
		case strings.HasPrefix(version, "0.6"):
			return nil
		default:
			migrator := stormv05.NewMigrator(b, m.forceCodec)
			err = migrator.Run(m.instances, m.kvKeys)
		}

		if err != nil {
			return err
		}
	}
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

func (m *Migrator) getVersion(b *bolt.DB) (string, error) {
	db, err := stormv05.Open("", stormv05.UseDB(b))
	if err != nil {
		return "", err
	}

	var v string
	_ = db.Get("__storm_db", "version", &v)

	return v, nil
}

// Codec option forces the codec used for the whole migration
func Codec(codec codec.MarshalUnmarshaler) func(*Migrator) error {
	return func(m *Migrator) error {
		m.forceCodec = codec
		return nil
	}
}
