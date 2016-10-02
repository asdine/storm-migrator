package json

import (
	"testing"

	"github.com/asdine/storm-migrator/v0.4/codec/internal"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
