package storage

import (
  _"fmt"
  _"log"

  "github.com/bwmarrin/discordgo"
	"github.com/ostafen/clover"
)



// represents a database
type Database struct {
  db *clover.DB
  isOpen bool
}



// initialize the database of a given toy. Requires the list of guilds the bot is in
func (d *Database) Open(guilds []*discordgo.Guild) error {
  if d.isOpen { return ErrOpened }

  // open the db
  db, err := clover.Open(StorageDirectory)
  if err != nil { return err }
  d.db = db
  d.isOpen = true


  // init/clean each collection
  for _, name := range collections {
    err = d.initCollection(name, guilds)
    if err != nil { return err }
  }

  return nil
}



// initialize and clean collection
func (d *Database) initCollection(
  name collectionName,
  sessionGuilds []*discordgo.Guild,
) (
  error,
) {

  // create the collection if it doesn't exist
  exists, _ := d.db.HasCollection(string(name))
  if !exists {
    _ = d.db.CreateCollection(string(name))
    //s.db.CreateIndex(collectionName, FieldGuildID)
    return nil // this skips the cleanup step kuz theres nothing to clean
  }

  // get collection
  query := d.db.Query(string(name))
  docs, _ := query.FindAll()

  // loop thru all docs
  for _, doc := range docs {
    shouldDelete := true
    guildID := doc.Get(FieldGuildID).(string)

    // remove entries if bot is not in guild anymore
    for _, g := range sessionGuilds {
      if guildID == g.ID {
        shouldDelete = false
        break
      }
    }

    if shouldDelete {
      err := query.DeleteById(doc.ObjectId())
      if err != nil { return err }
    }
  }

  return nil
}
