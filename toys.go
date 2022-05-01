package bubble

import (
  "fmt"
)

type Toy interface {
  Load(*Bot) error
  OnLifecycleEvent(LifecycleEvent) error

  ToyID() string
  ToyInfo() ToyInfo
}

type ToyInfo struct {
  Name, Description string
}


// loads/registers the toys provided in Config
func (b *Bot) loadToys(toys []Toy) error {
	b.toys = toys

	// register toys
	for _, e := range b.toys {
		if _, exists := b.toysByID[e.ToyID()]; exists {
			return fmt.Errorf(
        "[bot] Tried to load multiple toys with the ID %q", e.ToyID())
		}
		b.toysByID[e.ToyID()] = e
	}

	// now load them (we do this seperate so that each one has access to the
	// complete list of registered toys)
	for _, e := range b.toys {
		if err := e.Load(b); err != nil {
			return fmt.Errorf("[bot] Error loading extension: [%s] %w", e.ToyID(), err)
		}
	}

	return nil
}
