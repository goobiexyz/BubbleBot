package bubble

import (
	"fmt"
	"strings"
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	CoreVersion = "0.3"

	CoreName = "BubbleBot"
	DefaultName = "BubbleBot"
	FmtStatus = "%s v%s"
)


type Bot struct {
	Session *discordgo.Session

	name string
	version string
	onlineStatus string
	toys []Toy
	toysByID map[string]Toy
	storage *storage
	*msgManager
}

func (b *Bot) Name() string { return b.name }

func (b *Bot) UserID() string { return b.Session.State.User.ID }

func (b *Bot) Toys() []Toy { return b.toys }

func (b *Bot) FindToy(id string) (Toy, bool) {
	t, ok := b.toysByID[id]
	return t, ok
}


type Config struct {
	Name string
	Version string
	Token string
	Toys []Toy
	HideTimestamps bool
}


func NewBot(conf Config) (b *Bot, err error) {
	if conf.HideTimestamps { log.SetFlags(0) }

	// initialize Bot struct
	b = &Bot{
		toysByID		: make(map[string]Toy),
		storage     : new(storage),
	}

	// create discordgo session
	session, err := discordgo.New( "Bot " + conf.Token )
	if err != nil {
		 err = fmt.Errorf("error creating discordgo instance: %w", err)
		 return
	}
	b.Session = session

	// Set the discord api intents
	b.Session.Identify.Intents += discordgo.IntentsGuildMessages
	b.Session.Identify.Intents += discordgo.IntentsGuildMembers
	b.Session.Identify.Intents += discordgo.IntentsGuildPresences

	// initialize message manager
	b.msgManager = newMsgManager(b.Session)

	// set name, use default if none provided
	if name := strings.TrimSpace(conf.Name); name != "" {
		b.name = name
	} else {
		b.name = DefaultName
	}

	// set version
	b.version = strings.TrimSpace(conf.Version)

	// set the status that will show when bot is online
	b.onlineStatus = fmt.Sprintf(FmtStatus, CoreName, CoreVersion)
	if b.name != DefaultName && b.version != "" {
		b.onlineStatus += " | " + fmt.Sprintf(FmtStatus, b.name, b.version)
	}

	// register toys
	err = b.registerToys(conf.Toys)
	if err != nil { err = fmt.Errorf("error registering toys: %w", err) }

	return
}
