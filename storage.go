package bubble

import (
  "errors"
  _"fmt"
  "log"

  "github.com/bwmarrin/discordgo"
	"github.com/ostafen/clover"
  "golang.org/x/exp/slices"
)

var _ = log.Print // debugging

type collectionName string
const (
  CollectionGuilds collectionName   = "guilds"
  CollectionMembers collectionName  = "members"
  CollectionChannels collectionName = "channels"
  CollectionMessages collectionName = "messages"
  CollectionCustom collectionName = "custom"

  FieldGuildID        = "guildID"
  FieldUserID         = "memberID"
  FieldChannelID      = "channelID"
  FieldMessageID      = "messageID"
  FieldToyID          = "toyID"

  StorageDirectory    = "storage"
)

var (
  protectedFields = []string{ FieldGuildID, FieldUserID, FieldChannelID, FieldMessageID, FieldToyID }
  collections = []collectionName{
    CollectionGuilds,
    CollectionMembers,
    CollectionChannels,
    CollectionMessages,
    CollectionCustom,
  }

  ErrOpened error = errors.New("Database already open")
  ErrEntryExists error = errors.New("Entry already exists")
  ErrProtectedField error = errors.New("Cannot modify a protected field")
  ErrInvalidEntryType error = errors.New("Invalid entry type")
)


type storage struct {
  db *clover.DB
  open bool
}


// initialize the database of a given toy
func (s *storage) initDB(session *discordgo.Session) error {
  if s.open { return ErrOpened }

  // open the db
  db, err := clover.Open(StorageDirectory)
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
func (s *storage) initCollection(
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


// represents a set of methods for reading/writing to a specific database,
// constrained by certain rules
// FIXME: add system to prevent field name conflicts
type StorageDriver struct {
  *storage
}


type CustomEntry interface {
  GuildID() string
  ToyID() string
}


// creates a new document, adds it to the database, and returns it wrapped in an Entry
func (s *StorageDriver) createEntry(
  guildID string,
  collection collectionName,
  item interface{},
) (
  entry Entry,
  err error,
) {

  var doc *clover.Document
  if item != nil {
    doc = clover.NewDocumentOf(item)
    if doc == nil {
      err = ErrInvalidEntryType
      return
    }
  } else {
    doc = clover.NewDocument()
  }

  doc.Set(FieldGuildID, guildID)
  entry, err = s.addDocAndGetEntry(doc, collection)
  return
}


// adds the document to the database and returns the Entry wrapper for it
func (s *StorageDriver) addDocAndGetEntry(
  doc *clover.Document,
  collection collectionName,
) (
  entry Entry,
  err error,
) {

  id, err := s.db.InsertOne(string(collection), doc)
  if err != nil { return }

  doc, err = s.db.Query(string(collection)).FindById(id)
  if err != nil { return }

  entry = Entry{ doc, collection }
  return
}


func (s *StorageDriver) InsertMember(m *discordgo.Member) (e Entry, err error) {
  return s.InsertOne(m)
}


// insert a discord api object into the database
func (s *StorageDriver) InsertOne(item interface{}) (entry Entry, err error) {
  var criteria *clover.Criteria
  var collection collectionName
  var guildID string
  fields := make(map[string]interface{})

  switch i := any(item).(type) {
  case *discordgo.Guild:
    collection = CollectionGuilds
    guildID = i.ID
    criteria = clover.Field(FieldGuildID).Eq(i.ID)

  case *discordgo.Message:
    collection = CollectionMessages
    guildID = i.GuildID
    fields[FieldChannelID] = i.ChannelID
    fields[FieldMessageID] = i.ID
    criteria = clover.Field(FieldMessageID).Eq(i.ID)

  case *discordgo.Channel:
    collection = CollectionChannels
    guildID = i.GuildID
    fields[FieldChannelID] = i.ID
    criteria = clover.Field(FieldChannelID).Eq(i.ID)

  case *discordgo.Member:
    collection = CollectionMembers
    guildID = i.GuildID
    fields[FieldUserID] = i.User.ID
    criteria = clover.Field(FieldUserID).Eq(i.User.ID).And(clover.Field(FieldGuildID).Eq(i.GuildID))

  default:
    err = ErrInvalidEntryType
    return
  }

  query := s.db.Query(string(collection)).Where(criteria)
  if exists, _ := query.Exists(); exists {
    err = ErrEntryExists
    return
  }

  entry, err = s.createEntry(guildID, collection, fields)
  return
}


func (s *StorageDriver) CreateCustomEntry(
  guildID,
  toyID string,
  fields map[string]interface{},
) (
  entry Entry,
  err error,
) {

  fields[FieldToyID] = toyID
  entry, err = s.createEntry(guildID, CollectionCustom, fields)
  return
}


func (s *StorageDriver) Collection(name collectionName, filter map[string]interface{}) []Entry {
  query := s.db.Query(string(name))

  if filter != nil {
    for key, val := range filter {
      query = query.Where(clover.Field(key).Eq(val))
    }
  }

  docs, err := query.FindAll()
  if err != nil { panic(err) }
  entries := make([]Entry, len(docs))
  for i, doc := range docs {
    entries[i] = Entry{ doc, name }
  }
  return entries
}


// returns a guild Entry with the provided id
func (s *StorageDriver) Guild(id string) (Entry, error) {
  criteria := clover.Field(FieldGuildID).Eq(id)
  query := s.db.Query(string(CollectionGuilds)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionGuilds }, err
}

// returns a message Entry with the provided id
func (s *StorageDriver) Message(id string) (Entry, error) {
  criteria := clover.Field(FieldMessageID).Eq(id)
  query := s.db.Query(string(CollectionMessages)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionMessages }, err
}

// returns a channel Entry with the provided id
func (s *StorageDriver) Channel(id string) (Entry, error) {
  criteria := clover.Field(FieldChannelID).Eq(id)
  query := s.db.Query(string(CollectionChannels)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionChannels }, err
}

// returns a member Entry with the provided guild ID and user ID
func (s *StorageDriver) Member(guildID, userID string) (Entry, error) {
  criteria := clover.Field(FieldUserID).Eq(userID).And(clover.Field(FieldGuildID).Eq(guildID))
  query := s.db.Query(string(CollectionMembers)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionMembers }, err
}

// returns an entry from the custom collection, using an optional filter
func (s *StorageDriver) CustomEntry(toyID, guildID string, filter map[string]interface{}) (Entry, error) {
  criteria := clover.Field(FieldToyID).Eq(toyID).And(clover.Field(FieldGuildID).Eq(guildID))

  if filter != nil {
    for key, val := range filter {
      criteria = criteria.And(clover.Field(key).Eq(val))
    }
  }

  query := s.db.Query(string(CollectionCustom)).Where(criteria)
  doc, err := query.FindFirst()
  return Entry{ doc, CollectionCustom }, err
}



// returns if the database has an Entry for the guild with the given id
func (s *StorageDriver) HasGuild(id string) bool {
  criteria := clover.Field(FieldGuildID).Eq(id)
  query := s.db.Query(string(CollectionGuilds)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has an Entry for the message with the given id
func (s *StorageDriver) HasMessage(id string) bool {
  criteria := clover.Field(FieldMessageID).Eq(id)
  query := s.db.Query(string(CollectionMessages)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has an Entry for the channel with the given id
func (s *StorageDriver) HasChannel(id string) bool {
  criteria := clover.Field(FieldChannelID).Eq(id)
  query := s.db.Query(string(CollectionChannels)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has an Entry for the member with the given guild ID and user ID
func (s *StorageDriver) HasMember(guildID, userID string) bool {
  criteria := clover.Field(FieldUserID).Eq(userID).And(clover.Field(FieldGuildID).Eq(guildID))
  query := s.db.Query(string(CollectionChannels)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}

// returns if the database has entry from the custom collection with the guild and toyID, using an optional filter
func (s *StorageDriver) HasCustomEntry(toyID, guildID string, filter map[string]interface{}) bool {
  criteria := clover.Field(FieldToyID).Eq(toyID).And(clover.Field(FieldGuildID).Eq(guildID))

  if filter != nil {
    for key, val := range filter {
      criteria = criteria.And(clover.Field(key).Eq(val))
    }
  }

  query := s.db.Query(string(CollectionCustom)).Where(criteria)
  exists, _ := query.Exists()
  return exists
}



// saves or updates an Entry in the database
func (s *StorageDriver) Save(e Entry) error {
  return s.db.Save(string(e.collection), e.doc)
}

// removes an Entry from the database
// FIXME: later maybe implement some sort of garbage collection thing where
// toys can declare that they aren't using an entry anymore, and its only deleted
// when all the toys declare that ??
func (s *StorageDriver) Delete(e Entry) error {
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
