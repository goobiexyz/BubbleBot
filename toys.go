package bubble

import (
  "fmt"
  _"log"
)

type Toy interface {
  Load(*Bot, *Storage) error
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

  b.toyStores = make([]*Storage, len(b.toys))
	for i, t := range b.toys {
    s := &Storage{ name: t.ToyID() }
    b.toyStores[i] = s
    err := t.Load(b, s)

    if err != nil {
      Log(Error, b.name, fmt.Sprintf("Failed to load %q toy: %w", t.ToyID(), err))
      continue
    }
	}
}
