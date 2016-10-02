package gob

import (
	"testing"

	"github.com/asdine/storm-migrator/v0.4/codec/internal"
)

func TestGob(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
