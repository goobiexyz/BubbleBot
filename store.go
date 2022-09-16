package bubble

import (
  "errors"
  "fmt"
  _"log"

  "github.com/bwmarrin/discordgo"
	"github.com/ostafen/clover"
  "golang.org/x/exp/slices"
)

type collectionName string
const (
  CollectionGuilds collectionName   = "guilds"
  CollectionMembers collectionName  = "members"
  CollectionChannels collectionName = "channels"
  CollectionMessages collectionName = "messages"

  FieldGuildID        = "guildID"
  FieldUserID         = "memberID"
  FieldChannelID      = "channelID"
  FieldMessageID      = "messageID"

  StorageDirectory    = "storage"
)

var (
  protectedFields = []string{ FieldGuildID, FieldUserID, FieldChannelID, FieldMessageID }
  collections = []collectionName{
    CollectionGuilds,
    CollectionMembers,
    CollectionChannels,
    CollectionMessages,
  }

  ErrOpened error = errors.New("Database already open")
  ErrEntryExists error = errors.New("Entry already exists")
  ErrProtectedField error = errors.New("Cannot modify a protected field")
)

// database with four collections:
// guilds, members, channels, messages, all of which are scanned through on
// startup and deleted if the bot is not in the guild they belong to. It is the
// responsibility of each toy to handle deleting records for deleted messages
// and channels and such, because it is unnecessary to check for deleted things
// at startup kuz it will cause slowdown having to make all those consecutive
// api requests when u could just delete an Entry for when ur requesting one
// that isn't there. AND ur not even gonna be doing that a lot anyways.

type Storage struct {
  name string
  db *clover.DB
  open bool
}

// all are indexed by guildID, kuz when ur querying data, you usually are looking
// for stuff that only pertains to one guild. An exception would be tracking messages
// for reaction roles, and in that case you should loop through the list of guilds
// the bot is in and then loop thru each message being tracked for it kuz I think
// that'll be faster kuz of the indexing.
// add new entries only when theres a query for one that doesnt exist yet.


// initialize the database of a given toy
func (s *Storage) initDB(session *discordgo.Session) error {
  if s.open { return ErrOpened }

  // open the db
  db, err := clover.Open(fmt.Sprintf("%s/%s", StorageDirectory, s.name))
  if err != nil { return err }
  s.open = true
  s.db = db

  // go thru init/clean process with each collection kuz just because one collection
  // has an Entry for a guild doesn't mean that they all have that guild in their
  // entries too
  for _, name := range collections {
    err = s.initCollection(name, session.State.Ready.Guilds)
    if err != nil { return err }
  }

  return nil
}


// initialize and clean collection
func (s *Storage) initCollection(
  name collectionName,
  sessionGuilds []*discordgo.Guild,
) (
  error,
) {

  // create the collection if it doesn't exist
  exists, _ := s.db.HasCollection(string(name))
  if !exists {
    _ = s.db.CreateCollection(string(name))
    //s.db.CreateIndex(collectionName, FieldGuildID)
    return nil // this skips the cleanup step kuz theres nothing to clean
  }

  // get collection
  query := s.db.Query(string(name))
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


// represents the kinds of entries that can be added to the database
type storable interface {
  *discordgo.Guild | *discordgo.Message | *discordgo.Channel | *discordgo.Member
}


// insert a new Entry into the database provided
func InsertOne[D storable](s *Storage, discordItem D) (e Entry, err error) {
  var criteria *clover.Criteria
  var collection collectionName
  doc := clover.NewDocument()

  switch d := any(discordItem).(type) {
  case *discordgo.Guild:
    doc.Set(FieldGuildID, d.ID)
    criteria = clover.Field(FieldGuildID).Eq(d.ID)
    collection = CollectionGuilds

  case *discordgo.Message:
    doc.Set(FieldGuildID, d.GuildID)
    doc.Set(FieldChannelID, d.ChannelID)
    doc.Set(FieldMessageID, d.ID)
    criteria = clover.Field(FieldMessageID).Eq(d.ID)
    collection = CollectionMessages

  case *discordgo.Channel:
    doc.Set(FieldGuildID, d.GuildID)
    doc.Set(FieldChannelID, d.ID)
    criteria = clover.Field(FieldChannelID).Eq(d.ID)
    collection = CollectionChannels

  case *discordgo.Member:
    doc.Set(FieldGuildID, d.GuildID)
    doc.Set(FieldUserID, d.User.ID)
    criteria = clover.Field(FieldUserID).Eq(d.User.ID).And(clover.Field(FieldGuildID).Eq(d.GuildID))
    collection = CollectionMembers
  }

  query := s.db.Query(string(collection)).Where(criteria)
  if exists, _ := query.Exists(); exists {
    err = ErrEntryExists
    return
  }

  id, err := s.db.InsertOne(string(collection), doc)
  if err != nil { return }

  doc, err = s.db.Query(string(collection)).FindById(id)
  if err != nil { return }

  e = Entry{ doc, collection }
  return
}


func (s *Storage) Collection(name collectionName) []Entry {
  docs, err := s.db.Query(string(name)).FindAll()
  if err != nil { panic(err) }
  entries := make([]Entry, len(docs))
  for i, doc := range docs {
    entries[i] = Entry{ doc, name }
  }
  return entries
}


// returns a guild Entry with the provided id
func (s *Storage) Guild(id string) (Entry, error) {
  criteria := clover.Field(FieldGuildID).Eq(id)
  query := s.db.Query(string(CollectionGuilds)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionGuilds }, err
}

// returns a message Entry with the provided id
func (s *Storage) Message(id string) (Entry, error) {
  criteria := clover.Field(FieldMessageID).Eq(id)
  query := s.db.Query(string(CollectionMessages)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionMessages }, err
}

// returns a channel Entry with the provided id
func (s *Storage) Channel(id string) (Entry, error) {
  criteria := clover.Field(FieldChannelID).Eq(id)
  query := s.db.Query(string(CollectionChannels)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionChannels }, err
}

// returns a member Entry with the provided guild ID and user ID
func (s *Storage) Member(guildID, userID string) (Entry, error) {
  criteria := clover.Field(FieldUserID).Eq(userID).And(clover.Field(FieldGuildID).Eq(guildID))
  query := s.db.Query(string(CollectionMessages)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionMessages }, err
}



// returns if the database has an Entry for the guild with the given id
func (s *Storage) HasGuild(id string) bool {
  criteria := clover.Field(FieldGuildID).Eq(id)
  query := s.db.Query(string(CollectionGuilds)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has an Entry for the message with the given id
func (s *Storage) HasMessage(id string) bool {
  criteria := clover.Field(FieldMessageID).Eq(id)
  query := s.db.Query(string(CollectionMessages)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has an Entry for the channel with the given id
func (s *Storage) HasChannel(id string) bool {
  criteria := clover.Field(FieldChannelID).Eq(id)
  query := s.db.Query(string(CollectionChannels)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has an Entry for the member with the given guild ID and user ID
func (s *Storage) HasMember(guildID, userID string) bool {
  criteria := clover.Field(FieldUserID).Eq(userID).And(clover.Field(FieldGuildID).Eq(guildID))
  query := s.db.Query(string(CollectionChannels)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}



// saves or updates an Entry in the database
func (s *Storage) Save(e Entry) error {
  return s.db.Save(string(e.collection), e.doc)
}

// removes an Entry from the database
func (s *Storage) Delete(e Entry) error {
  return s.db.Query(string(e.collection)).DeleteById(e.doc.ObjectId())
}


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
