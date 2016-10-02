package migrator_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	migrator "github.com/asdine/storm-migrator"
	stormv04 "github.com/asdine/storm-migrator/v0.4"
	stormv05 "github.com/asdine/storm-migrator/v0.5"
	"github.com/asdine/storm-migrator/v0.5/codec/json"
	"github.com/stretchr/testify/require"
)

func TestMigrator(t *testing.T) {
	dir, path, cleanup := prepareDB(t)
	defer cleanup()

	m := migrator.New(path)
	m.AddBuckets(new(A), new(B))
	m.AddKV(
		"bucket",
		[]interface{}{
			new(int),
			new(string),
		})
	err := m.Run(filepath.Join(dir, "v05.db"), migrator.Codec(json.Codec))
	require.NoError(t, err)

	db, err := stormv05.Open(filepath.Join(dir, "v05.db"))
	require.NoError(t, err)

	var AList []A
	err = db.All(&AList)
	require.NoError(t, err)

	var a A
	var b B
	for i := 0; i < 10; i++ {
		err = db.One("ID", i+1, &a)
		require.NoError(t, err)
		require.Equal(t, i+1, a.ID)
		require.Equal(t, fmt.Sprintf("Field%d", i), a.Field1)
		require.False(t, a.Field2.IsZero())

		err = db.One("ID", strconv.Itoa(i+1), &b)
		require.NoError(t, err)
		require.Equal(t, strconv.Itoa(i+1), b.ID)
		require.Equal(t, int64(i*10), b.Field1)

		var v int
		switch {
		case i%2 == 0:
			fmt.Println(fmt.Sprintf("string%d", i))
			err = db.Get("bucket", fmt.Sprintf("string%d", i), &v)
			require.NoError(t, err)
			require.Equal(t, i, v)

			err = db.Get("bucket", fmt.Sprintf("string%d", i+1), &a)
			require.NoError(t, err)
			require.Equal(t, i+11, a.ID)
		default:
			err = db.Get("bucket", i+11, &v)
			require.NoError(t, err)
			require.Equal(t, i, v)

			err = db.Get("bucket", i+12, &a)
			require.NoError(t, err)
			require.Equal(t, i+12, a.ID)
		}
	}
}

func prepareDB(t *testing.T) (string, string, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "storm-migrator")
	require.NoError(t, err)

	path := filepath.Join(dir, "v04.db")
	dbv04, err := stormv04.Open(path)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err = dbv04.Save(&A{ID: i + 1, Field1: fmt.Sprintf("Field%d", i), Field2: time.Now()})
		require.NoError(t, err)
		err = dbv04.Save(&B{ID: strconv.Itoa(i + 1), Field1: int64(i * 10)})
		require.NoError(t, err)
		switch {
		case i%2 == 0:
			err = dbv04.Set("bucket", fmt.Sprintf("string%d", i), i)
			err = dbv04.Set("bucket", fmt.Sprintf("string%d", i+1), &A{ID: i + 11})
		default:
			err = dbv04.Set("bucket", i+11, i)
			err = dbv04.Set("bucket", i+12, &A{ID: i + 12})
		}
	}
	dbv04.Close()

	return dir, path, func() {
		os.RemoveAll(dir)
	}
}

type A struct {
	ID     int
	Field1 string
	Field2 time.Time
}

type B struct {
	ID     string
	Field1 int64
}
