package bubble

import (
	"fmt"
	"strings"
	"log"

	"github.com/bwmarrin/discordgo"
)


type Bot struct {
	name string
	Session *discordgo.Session
	toys []Toy
	toysByID map[string]Toy
	*msgManager
}

func (b *Bot) UserID() string { return b.Session.State.User.ID }

func (b *Bot) Toys() []Toy { return b.toys }

func (b *Bot) FindToy(id string) (Toy, bool) {
	t, ok := b.toysByID[id]
	return t, ok
}


type Config struct {
	Name string
	Token string
	Toys []Toy
	HideTimestamps bool
}


func NewBot(conf Config) (b *Bot, err error) {
	if conf.HideTimestamps { log.SetFlags(0) }

	// create discordgo session
	dg, err := discordgo.New( "Bot " + conf.Token )
	if err != nil {
		 err = fmt.Errorf("error creating discordgo instance: %w", err)
		 return
	}

	// We only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// default name is BubbleBot if none is provided
	name := strings.TrimSpace(conf.Name)
	if name == "" { name = "BubbleBot" }

	// create Bot struct
	b = &Bot{
		name 				: name,
		Session			: dg,
		toysByID		: make(map[string]Toy),
		msgManager  : newMsgManager(dg),
	}

	// register toys
	err = b.registerToys(conf.Toys)
	if err != nil { err = fmt.Errorf("error registering toys: %w", err) }

	return
}
