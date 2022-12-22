package storage

import (
  _"fmt"
  _"log"

	"github.com/ostafen/clover"
)




// represents a set of methods for reading/writing to the database,
// constrained by certain rules. This is given to each Toy.
// FIXME: add system to prevent field name conflicts
type StorageDriver struct {
  *Database
}



// Creates a new document, adds it to the database, and returns it wrapped in an Entry
// Also you can provide a *clover.Criteria to stop the creation of the document if something
// already satisfies that query
func (s *StorageDriver) createEntry(
  guildID string,
  collection collectionName,
  item interface{},
  criteria *clover.Criteria,
) (
  entry Entry,
  err error,
) {

  // if a criteria is provided, an error will be returned if something already
  // satisfies that query
  if criteria != nil {
    query := s.db.Query(string(collection)).Where(criteria)
    if exists, _ := query.Exists(); exists {
      err = ErrEntryExists
      return
    }
  }

  // create the document from an object of arbitrary type, erroring if clover is
  // unable to convert it
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

  // set the guildID field to the guildID provided
  doc.Set(FieldGuildID, guildID)

  // add the doc to the database
  id, err := s.db.InsertOne(string(collection), doc)
  if err != nil { return }

  // query the doc to get it back
  doc, err = s.db.Query(string(collection)).FindById(id)
  if err != nil { return }

  // wrap it in an Entry object to return
  entry = Entry{ doc, collection }
  return
}





// ========== PUBLIC METHODS ==========





// returns a list of entries in a collection, using an optional filter
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








// Insert a member to the database
func (s *StorageDriver) InsertMember(guildID, userID string) (Entry, error) {
  fields := make(map[string]interface{})
  fields[FieldUserID] = userID

  criteria := clover.Field(FieldUserID).Eq(userID).And(clover.Field(FieldGuildID).Eq(guildID))

  return s.createEntry(
    guildID,
    CollectionMembers,
    fields,
    criteria,
  )
}



// Insert a guild to the database
func (s *StorageDriver) InsertGuild(guildID string) (Entry, error) {
  criteria := clover.Field(FieldGuildID).Eq(guildID)

  return s.createEntry(guildID, CollectionGuilds, make(map[string]interface{}), criteria)
}



// Insert a message to the database
func (s *StorageDriver) InsertMessage(guildID string, messageID string, channelID string) (Entry, error) {
  fields := make(map[string]interface{})
  fields[FieldChannelID] = channelID
  fields[FieldMessageID] = messageID

  criteria := clover.Field(FieldMessageID).Eq(messageID)

  return s.createEntry(guildID, CollectionMessages, fields, criteria)
}



// Insert a channel to the database
func (s *StorageDriver) InsertChannel(guildID string, channelID string) (Entry, error) {
  fields := make(map[string]interface{})
  fields[FieldChannelID] = channelID

  criteria := clover.Field(FieldChannelID).Eq(channelID)

  return s.createEntry(guildID, CollectionChannels, fields, criteria)
}



// Insert a custom entry to the database
func (s *StorageDriver) InsertCustomEntry(
  guildID string,
  toyID string,
  fields map[string]interface{},
) (
  Entry,
  error,
) {

  fields[FieldToyID] = toyID
  return s.createEntry(guildID, CollectionCustom, fields, nil)
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
