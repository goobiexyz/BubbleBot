package storage

import (
  "errors"
)



type collectionName string; const (
  CollectionGuilds collectionName = "guilds"
  CollectionMembers collectionName = "members"
  CollectionChannels collectionName = "channels"
  CollectionMessages collectionName = "messages"
  CollectionCustom collectionName = "custom"
)



const (
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
