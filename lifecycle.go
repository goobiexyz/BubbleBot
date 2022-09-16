package bubble

import (
	"os"
	"os/signal"
	"syscall"
  "time"
  "fmt"
)



type LifecycleEvent int
const (
	Connect LifecycleEvent = iota
	Close
)


// Start the bot
func (b *Bot) Start() {

  // Load toys
  b.loadToys()

  // Connect to Discord
  b.connect()

  // if connection was successful, ensure that the bot will
  // stop by the end of this function, even if there's a panic()
	defer b.Stop()

	// Wait here until CTRL-C or other term signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	Log(lifecycle, b.name, "Bot started successfully. Press Ctrl+C to stop")
	<-stop
  fmt.Println() // add a new line after the ^C thingy
}


// Stop the bot
func (b *Bot) Stop() {
  Log(Info, b.name, "Shutting down")

  for _, e := range b.toys { e.OnLifecycleEvent(Close) }

  b.close()

	Log(lifecycle, b.name, "Gracefully shut down")
}


// Open a websocket connection to Discord
func (b *Bot) connect() {
  // keep trying to connect
  for {
    Log(Info, b.name, "Connecting to Discord")
    err := b.Session.Open()
    if err == nil { break }

    Log(Error, b.name, "Couldn't open session: " + err.Error())
    Log(Info, b.name, "Will attempt to reconnect in 10 seconds")
    time.Sleep(10*time.Second)
  }

  Log(lifecycle, b.name, "Connected successfully")

	// Initialize the stores for the toys
	for _, s := range b.toyStores {
		err := s.initDB(b.Session)
		if err != nil { panic(err) }
	}

  // send connect event to toys
	for _, t := range b.toys { t.OnLifecycleEvent(Connect) }
}


// Close the connection
func (b *Bot) close() {
  Log(Info, b.name, "Closing connection")
  err := b.Session.Close()
  if err != nil {
    Log(Error, b.name, "Failed to close connection: " + err.Error())
  }
}
