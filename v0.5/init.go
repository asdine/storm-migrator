package storm

import (
	"reflect"

	"github.com/asdine/storm-migrator/v0.5/index"
	"github.com/boltdb/bolt"
)

// Init creates the indexes and buckets for a given structure
func (n *node) Init(data interface{}) error {
	v := reflect.ValueOf(data)
	info, err := extract(&v)
	if err != nil {
		return err
	}

	return n.readWriteTx(func(tx *bolt.Tx) error {
		return n.init(tx, info)
	})
}

func (n *node) init(tx *bolt.Tx, info *modelInfo) error {
	bucket, err := n.CreateBucketIfNotExists(tx, info.Name)
	if err != nil {
		return err
	}

	// save node configuration in the bucket
	err = n.saveMetadata(bucket)
	if err != nil {
		return err
	}

	for fieldName, idxInfo := range info.Indexes {
		switch idxInfo.Type {
		case tagUniqueIdx:
			_, err = index.NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
		case tagIdx:
			_, err = index.NewListIndex(bucket, []byte(indexPrefix+fieldName))
		default:
			err = ErrIdxNotFound
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Init creates the indexes and buckets for a given structure
func (s *DB) Init(data interface{}) error {
	return s.root.Init(data)
}
