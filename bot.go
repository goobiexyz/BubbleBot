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


func NewBot(conf Config) (*Bot, error) {
	if conf.HideTimestamps { log.SetFlags(0) }

	dg, err := discordgo.New( "Bot " + conf.Token )
	if err != nil {
		 return nil, fmt.Errorf("[discordgo.New] %w", err)
	}

	// We only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	name := strings.TrimSpace(conf.Name)
	if name == "" { name = "BubbleBot" }

	b := &Bot{
		name 				: name,
		Session			: dg,
		toysByID		: make(map[string]Toy),
		msgManager  : newMsgManager(dg),
	}

	if err := b.loadToys(conf.Toys); err != nil { return nil, err }

	return b, nil
}
