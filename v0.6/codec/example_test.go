package codec_test

import (
	"fmt"

	"github.com/asdine/storm-migrator/v0.6"
	"github.com/asdine/storm-migrator/v0.6/codec/gob"
	"github.com/asdine/storm-migrator/v0.6/codec/json"
	"github.com/asdine/storm-migrator/v0.6/codec/protobuf"
	"github.com/asdine/storm-migrator/v0.6/codec/sereal"
)

func Example() {
	// The examples below show how to set up all the codecs shipped with Storm.
	// Proper error handling left out to make it simple.
	var gobDb, _ = storm.Open("gob.db", storm.Codec(gob.Codec))
	var jsonDb, _ = storm.Open("json.db", storm.Codec(json.Codec))
	var serealDb, _ = storm.Open("sereal.db", storm.Codec(sereal.Codec))
	var protobufDb, _ = storm.Open("protobuf.db", storm.Codec(protobuf.Codec))

	fmt.Printf("%T\n", gobDb.Codec())
	fmt.Printf("%T\n", jsonDb.Codec())
	fmt.Printf("%T\n", serealDb.Codec())
	fmt.Printf("%T\n", protobufDb.Codec())

	// Output:
	// *gob.gobCodec
	// *json.jsonCodec
	// *sereal.serealCodec
	// *protobuf.protobufCodec
}
