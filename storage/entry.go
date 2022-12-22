package storage

import (
  _"fmt"
  _"log"

	"github.com/ostafen/clover"
  "golang.org/x/exp/slices"
)




// represents an Entry in the database
type Entry struct {
  doc *clover.Document
  collection collectionName
}



// returns the value of the field provided
func (e Entry) Get(field string) interface{} {
  return e.doc.Get(field)
}



// tries to set the value of the field provided
func (e Entry) Set(field string, val interface{}) error {
  if slices.Contains(protectedFields, field) { return ErrProtectedField }
  e.doc.Set(field, val)
  return nil
}
