package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	stormv05 "github.com/asdine/storm-migrator/v0.5"
	"github.com/asdine/storm-migrator/v0.6/codec/json"
	"github.com/stretchr/testify/require"
)

func TestMigrator(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	dbv05, err := stormv05.Open(filepath.Join(dir, "my.db"))
	require.NoError(t, err)
	defer dbv05.Close()

	type User struct {
		ID              int       `storm:"id"`
		Name            string    `storm:"index"`
		Age             int       `storm:"index"`
		DateOfBirth     time.Time `storm:"index"`
		Group           string
		unexportedField int
		Slug            string `storm:"unique"`
	}

	for i := 0; i < 10; i++ {
		err = dbv05.Save(&User{ID: i + 1, Name: "John"})
		require.NoError(t, err)
	}

	m := NewMigrator(dbv05.Bolt, json.Codec)
	err = m.Run([]interface{}{new(User)}, nil)
	require.NoError(t, err)

	dbv06, err := Open("", UseDB(dbv05.Bolt))
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		var u User
		err = dbv06.One("ID", i+1, &u)
		require.NoError(t, err)
		require.Equal(t, i+1, u.ID)
	}
}
