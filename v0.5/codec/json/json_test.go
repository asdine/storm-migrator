package json

import (
	"testing"

	"github.com/asdine/storm-migrator/v0.5/codec/internal"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
