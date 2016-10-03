# Storm Migrator

This library can be used to migrate existing databases to newest version of [Storm](https://github.com/asdine/storm).

## Getting started

Get the latest version of Storm

```
$ go get -u github/asdine/storm
```

Install the migrator

```
$ go get -u github.com/asdine/storm-migrator
```

The migration must be done manually on the database of your choice.
The migrator **does not** overwrite your existing database, it creates a copy and operates on it.

```go
package main

import (
	"log"

	migrator "github.com/asdine/storm-migrator"
)

func main() {
  // Instantiate a migrator and select the database
  m := migrator.New("/path/to/my.db")

  // There are two types of buckets created by Storm:
  //    those who were created by db.Save or db.Init
  //    those who were created by db.Set

  // If you have buckets created by db.Save or db.Init
  // You need to give a new instance of every type
  m.AddBuckets(new(User), new(Project), new(mypackage.MyType))

  // If you have buckets created by db.Set
  // You need to give an instance of every type used for the KEY
	m.AddKV("the-bucket", []interface{}{
    new(int),
    new(map[.....]...)
    new([]byte), // always make sure that []byte and string are the last elements
    new(string),
  })

  // Then run by specifying the path of the new database.
  // The database MUST be a non existing file.
  // The codec MUST be passed
	err := m.Run("new.db")
	if err != nil {
		log.Fatal(err)
	}
}
```

Alternatively, the codec can be passed to `Run`

```go
m.Run("new.db", migrator.Codec(json.Codec))
```

## Issues

Don't hesitate opening an issue if the migration doesn't work as expected
