package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	stormv04 "github.com/asdine/storm-migrator/v0.4"
	"github.com/asdine/storm-migrator/v0.6/codec/json"
	"github.com/stretchr/testify/require"
)

func TestMigrator(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	dbv04, err := stormv04.Open(filepath.Join(dir, "my.db"))
	require.NoError(t, err)
	defer dbv04.Close()

	for i := 0; i < 10; i++ {
		err = dbv04.Save(&User{ID: i + 1, Name: "John"})
		require.NoError(t, err)
	}

	err = dbv04.Set("bucket", "string", "value")
	require.NoError(t, err)
	err = dbv04.Set("bucket", 1, "value")
	require.NoError(t, err)
	err = dbv04.Set("bucket", []byte("a slice of bytes"), "value")
	require.NoError(t, err)
	err = dbv04.Set("bucket", User{ID: 10}, "value")
	require.NoError(t, err)

	m := NewMigrator(dbv04.Bolt, json.Codec)
	err = m.Run([]interface{}{new(User)}, map[string][]interface{}{
		"bucket": {new(User), new(int), new([]byte), new(string)},
	})
	require.NoError(t, err)

	dbv05, err := Open("", UseDB(dbv04.Bolt))
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		var u User
		err = dbv05.One("ID", i+1, &u)
		require.NoError(t, err)
		require.Equal(t, i+1, u.ID)
	}

	var s string
	err = dbv05.Get("bucket", "string", &s)
	require.NoError(t, err)
	err = dbv05.Get("bucket", 1, &s)
	require.NoError(t, err)
	err = dbv05.Get("bucket", []byte("a slice of bytes"), &s)
	require.NoError(t, err)
	err = dbv05.Get("bucket", User{ID: 10}, &s)
	require.NoError(t, err)

}
