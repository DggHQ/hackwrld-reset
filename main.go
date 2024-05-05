package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/DggHQ/hackwrld-reset/bot"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Msg struct {
	Data string `json:"data"`
}

var (
	valkeyHost  = getEnv("VALKEY_HOST", "valkey.hackwrld.svc")
	valkeyctx   = context.Background()
	restartTime = time.Now().Add(time.Minute * 30)
	u           = url.URL{
		Scheme:   getEnv("SCHEME", "ws"),
		Host:     fmt.Sprintf("%s:%s", getEnv("HOST", "localhost"), getEnv("PORT", "8080")),
		Path:     "/ws",
		RawQuery: fmt.Sprintf("token=%s", getEnv("KEY", "secret")),
	}
	wsmessage = make(chan []byte)
	chatmsg   = make(chan string)
	minutes   = 30
	operation = getEnv("OPERATION", "DEPLOYMENTS")
	chatKey   = getEnv("CHATKEY", "NONE")
	chatBot   = bot.Bot{}
)

// Handle setting of variables of env var is not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// This keeps the webcocket connection alive. The server handles each client connection.
func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			log.Println(err)
			c.Close()
			break
		}
	}
}

func main() {

	switch operation {
	case "DEPLOYMENTS":
		log.Println("Waiting 30 minutes until game reset.")
		// Connect to websocket to send messages every 5 minutes
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		go readLoop(c)
		go chatBot.Start(chatKey, chatmsg)
		// Launch goroutine that handles the messages that come into the channel and write them to the connection.
		go func(connection *websocket.Conn) {
			for msg := range wsmessage {
				ok := c.WriteMessage(websocket.TextMessage, msg)
				if ok != nil {
					log.Println("write:", ok)
				}
			}
		}(c)

		for {
			if time.Now().Before(restartTime) {
				log.Print("Waiting for restart. Letting players know")
				reminder := fmt.Sprintf("[HACKWRLD will start a new game in %d minutes]", minutes)
				msg := Msg{
					Data: reminder,
				}
				minutes = minutes - 10
				message, err := json.Marshal(msg)
				if err != nil {
					log.Println(err)
				}
				// Write message to channel to be written to websocket connection
				wsmessage <- message
				// Write to DGG Chat
				chatmsg <- reminder
				// Sleep
				time.Sleep(time.Minute * 10)
				continue
			}
			log.Println("Timer done. Deleting current game")
			// Let players know game restarts
			reminder := "[HACKWRLD will now start a new game. Please login here if you wanna play: https://hackwrld.notacult.website/]"
			msg := Msg{
				Data: reminder,
			}
			message, err := json.Marshal(msg)
			if err != nil {
				log.Println(err)
			}
			wsmessage <- message
			chatmsg <- reminder
			// Set the current timestamp as a key in valkey to get for the leaderboard static time values
			rdb := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:6379", valkeyHost),
				Password: "", // no password set
				DB:       0,  // use default DB
			})
			timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
			err = rdb.Set(valkeyctx, "ts", timestamp[0:13], 0).Err()
			if err != nil {
				log.Println(err)
			}
			os.Exit(0)
		}
	case "STATE":
		os.Exit(0)
	}

}
