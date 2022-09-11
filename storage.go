package bubble

import (
  "fmt"
  _"log"

  "github.com/bwmarrin/discordgo"
	"github.com/ostafen/clover"
)


type Storage struct {
  db *clover.DB
}


// opens the database
func (s *Storage) InitDB(session *discordgo.Session) error {
  // open the clover database
  db, err := clover.Open("storage")
  if err != nil { return err }
  s.db = db

  // make sure it has a collection called guild_options
  hasGuildOptions, _ := s.db.HasCollection("guild_options")
  if !hasGuildOptions {
    _ = s.db.CreateCollection("guild_options")
  }

  // remove docs for guilds that the bot isn't in anymore
  s.clean(session.State.Ready.Guilds)

  // add handler for when bot is added to guild, add it to database
  session.AddHandler(func (_ *discordgo.Session, g *discordgo.GuildCreate) {
    s.addGuild(g.Guild.ID)
  })

  // add handler for when bot is added to guild, remove it from database
  session.AddHandler(func (_ *discordgo.Session, g *discordgo.GuildDelete) {
    s.deleteGuild(g.Guild.ID)
  })

  return nil
}


// add a document to represent a guild in the database
func (s *Storage) addGuild(guildID string) {
  query := s.db.Query("guild_options").Where(clover.Field("id").Eq(guildID))
  if exists, _ := query.Exists(); exists { return }

  doc := clover.NewDocument()
  doc.Set("id", guildID)
  _ = s.db.Insert("guild_options", doc)
}


// remove a guild document from the database
func (s *Storage) deleteGuild(guildID string) {
  query := s.db.Query("guild_options").Where(clover.Field("id").Eq(guildID))
  if exists, _ := query.Exists(); exists { _ = query.Delete() }
}


// sync up the list of guilds with the guilds the bot is in
func (s *Storage) clean(guilds []*discordgo.Guild) {
  query := s.db.Query("guild_options")
  guildDocs, _ := query.FindAll()

  // remove docs for guilds that the bot isnt in anymore
  for _, guildDoc := range guildDocs {
    shouldDelete := true
    for _, g := range guilds {
      if guildDoc.Get("id").(string) == g.ID {
        shouldDelete = false
        break
      }
    }

    if shouldDelete { _ = query.DeleteById(guildDoc.ObjectId()) }
  }

  // make sure theres a doc for every guild the bot is in
  for _, g := range guilds { s.addGuild(g.ID) }
}


// get the value of a guild option
func (s *Storage) Get(guildID, name string) (interface{}, error) {
  criteria := clover.Field("id").Eq(guildID)
  doc, _ := s.db.Query("guild_options").Where(criteria).FindFirst()

  if doc == nil {
    return nil, fmt.Errorf("no database entry for guild with ID %q", guildID)
  }
  return doc.Get(name), nil
}


// set the value of a guild option
func (s *Storage) Set(guildID, name string, val interface{}) error {
  query := s.db.Query("guild_options").Where(clover.Field("id").Eq(guildID))

  if e, _ := query.Exists(); !e {
    return fmt.Errorf("no database entry for guild with ID %q", guildID)
  }

  query.Update(map[string]interface{}{ name : val })
  return nil
}
