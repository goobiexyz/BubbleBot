package bubble


import (
  "github.com/bwmarrin/discordgo"
)


// must return true if it took any action on the message,
// so that any message handlers afterwards will not be run.
// this is to prevent conflicts
type MsgHandler func(*discordgo.MessageCreate) bool


type msgManager struct {
  handlers []MsgHandler
}


func newMsgManager(s *discordgo.Session) *msgManager {
  man := &msgManager{}
  s.AddHandler(man.onMessage)
  return man
}


func (man *msgManager) onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID { return }

	// run through message handlers. Break if one of them did something with
  // the message, to prevent conflicts.
	for _, handle := range man.handlers {
    if handle(m) { break }
  }
}


func (man *msgManager) AddMsgHandler(h MsgHandler) {
	man.handlers = append(man.handlers, h)
}
