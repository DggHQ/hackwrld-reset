package bot

import (
	"log"
	"time"

	"github.com/MemeLabs/dggchat"
)

var (
	lastPing = timeToUnix(time.Now())
	lastPong = timeToUnix(time.Now())
)

const (
	pingInterval = time.Minute
)

type Bot struct {
	Session *dggchat.Session
	Message chan string
}

func (b *Bot) Start(key string, messagechan chan string) {
	dgg, err := dggchat.New(key)
	if err != nil {
		log.Println(err)
	}
	b.Session = dgg
	b.Message = messagechan
	err = b.Session.Open()
	if err != nil {
		log.Println(err)
	}
	defer b.Session.Close()

	errors := make(chan string)
	pings := make(chan dggchat.Ping)

	b.Session.AddErrorHandler(func(e string, s *dggchat.Session) {
		errors <- e
	})

	b.Session.AddPingHandler(func(p dggchat.Ping, s *dggchat.Session) {
		pings <- p
	})
	// Reply to pings from the websocket
	go checkConnection(b.Session)
	go b.SendMessage()
	for {
		select {
		case e := <-errors:
			log.Printf("Error %s\n", e)
		case p := <-pings:
			lastPong = p.Timestamp
		}
	}
}

// Send message to chat
func (b *Bot) SendMessage() {
	for msg := range b.Message {
		err := b.Session.SendMessage(msg)
		if err != nil {
			log.Println(err)
		}
	}

}

// Periodically ping websocket for keepalive
func checkConnection(s *dggchat.Session) {
	ticker := time.NewTicker(pingInterval)
	for {
		<-ticker.C
		if lastPing != lastPong {
			log.Println("Ping mismatch, attempting to reconnect")
			err := s.Close()
			if err != nil {
				log.Println(err)
			}

			err = s.Open()
			if err != nil {
				log.Println(err)
			}

			continue
		}
		s.SendPing()
		lastPing = timeToUnix(time.Now())
	}
}

func timeToUnix(t time.Time) int64 {
	return t.Unix() * 1000
}
