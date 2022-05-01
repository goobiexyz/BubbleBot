package bubble

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)



type LifecycleEvent int
const (
	Connect LifecycleEvent = iota
	Close
)


// Start the bot
func (b *Bot) Start() {
  Log(Info, b.name, "Connecting to Discord")

  // Connect to Discord
  err := b.connect()
  if err != nil { Log(Error, b.name, "Error on connect event: " + err.Error()) }

	defer b.Stop()

	// Wait here until CTRL-C or other term signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	Log(lifecycle, b.name, "Now running. Press Ctrl+C to stop")
	<-stop
}


// Stop the bot
func (b *Bot) Stop() {
  Log(Info, b.name, "Shutting down")

  err := b.close()
  if err != nil { Log(Error, b.name, "Error on close event: " + err.Error()) }

	Log(lifecycle, b.name, "Gracefully shut down")
}


// Open a websocket connection to Discord
func (b *Bot) connect() error {
	err := b.Session.Open()
	if err != nil { return fmt.Errorf("Couldn't open session: %w", err) }

	for _, e := range b.toys {
		err = e.OnLifecycleEvent(Connect)
		if err != nil { return fmt.Errorf( LogStr(Error, e.ToyID(), err.Error()) )
    }
	}

  return nil
}


// Close the connection
func (b *Bot) close() error {
  defer b.Session.Close()

  Log(Info, b.name, "Closing connection")

  for _, e := range b.toys {
		err := e.OnLifecycleEvent(Close)
    if err != nil { return fmt.Errorf( LogStr(Error, e.ToyID(), err.Error()) )
    }
	}

  return nil
}
