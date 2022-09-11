package bubble

import (
  "fmt"
  _"log"
)

type Toy interface {
  Load(*Bot) error
  OnLifecycleEvent(LifecycleEvent)

  ToyID() string
  ToyInfo() ToyInfo
}

type ToyInfo struct {
  Name, Description string
}


// registers the toys provided, making sure there's no duplicates
func (b *Bot) registerToys(toys []Toy) error {
	for _, t := range toys {
		if _, exists := b.toysByID[t.ToyID()]; exists {
			return fmt.Errorf("tried to register multiple toys with the ID %q", t.ToyID())
		}
    b.toys = append(b.toys, t)
		b.toysByID[t.ToyID()] = t
	}

  return nil
}



// calls the load function for each registered toy
func (b *Bot) loadToys() {
  Log(Info, b.name, "Loading toys")

	for _, t := range b.toys {
    err := t.Load(b)

    if err != nil {
      Log(Error, b.name, fmt.Sprintf("Failed to load %q toy: %w", t.ToyID(), err))
      continue
    }
	}
}


func (b *Bot) Option(toy Toy, guildID, name string) (interface{}, error) {
  optionName := toy.ToyID() + "_" + name
  return b.storage.Get(guildID, optionName)
}


func (b *Bot) SetOption(toy Toy, guildID, name string, val interface{}) error {
  optionName := toy.ToyID() + "_" + name
  return b.storage.Set(guildID, optionName, val)
}
