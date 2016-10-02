package storm

import (
	"os"
	"time"

	"github.com/asdine/storm-migrator/v0.4/codec"
	"github.com/asdine/storm-migrator/v0.4/codec/json"
	"github.com/boltdb/bolt"
)

const (
	metadataBucket = "__storm_metadata"
)

// Defaults to json
var defaultCodec = json.Codec

// Open opens a database at the given path with optional Storm options.
func Open(path string, stormOptions ...func(*DB) error) (*DB, error) {
	var err error

	s := &DB{
		Path:  path,
		codec: defaultCodec,
	}

	for _, option := range stormOptions {
		if err = option(s); err != nil {
			return nil, err
		}
	}

	if s.boltMode == 0 {
		s.boltMode = 0600
	}

	if s.boltOptions == nil {
		s.boltOptions = &bolt.Options{Timeout: 1 * time.Second}
	}

	s.root = &node{s: s, rootBucket: s.rootBucket, codec: s.codec}

	// skip if UseDB option is used
	if s.Bolt == nil {
		s.Bolt, err = bolt.Open(path, s.boltMode, s.boltOptions)
		if err != nil {
			return nil, err
		}

		err = s.checkVersion()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// DB is the wrapper around BoltDB. It contains an instance of BoltDB and uses it to perform all the
// needed operations
type DB struct {
	// Path of the database file
	Path string

	// Handles encoding and decoding of objects
	codec codec.EncodeDecoder

	// Bolt is still easily accessible
	Bolt *bolt.DB

	// Bolt file mode
	boltMode os.FileMode

	// Bolt options
	boltOptions *bolt.Options

	// Enable auto increment on empty integer fields
	autoIncrement bool

	// The root node that points to the root bucket.
	root *node

	// The root bucket name
	rootBucket []string
}

// From returns a new Storm node with a new bucket root.
// All DB operations on the new node will be executed relative to the given
// bucket.
func (s *DB) From(root ...string) Node {
	newNode := *s.root
	newNode.rootBucket = root
	return &newNode
}

// WithTransaction returns a New Storm node that will use the given transaction.
func (s *DB) WithTransaction(tx *bolt.Tx) Node {
	return s.root.WithTransaction(tx)
}

// Bucket returns the root bucket name as a slice.
// In the normal, simple case this will be empty.
func (s *DB) Bucket() []string {
	return s.root.Bucket()
}

// Close the database
func (s *DB) Close() error {
	return s.Bolt.Close()
}

// Codec returns the EncodeDecoder used by this instance of Storm
func (s *DB) Codec() codec.EncodeDecoder {
	return s.codec
}

// WithCodec returns a New Storm Node that will use the given Codec.
func (s *DB) WithCodec(codec codec.EncodeDecoder) Node {
	n := s.From().(*node)
	n.codec = codec
	return n
}

func (s *DB) checkVersion() error {
	var v string
	err := s.Get(metadataBucket, "version", &v)
	if err != nil && err != ErrNotFound {
		return err
	}

	// for now, we only set the current version if it doesn't exist
	if v == "" {
		return s.Set(metadataBucket, "version", Version)
	}

	return nil
}

// toBytes turns an interface into a slice of bytes
func toBytes(key interface{}, encoder codec.EncodeDecoder) ([]byte, error) {
	if key == nil {
		return nil, nil
	}
	if k, ok := key.([]byte); ok {
		return k, nil
	}
	if k, ok := key.(string); ok {
		return []byte(k), nil
	}

	return encoder.Encode(key)
}
