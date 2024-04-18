package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/DggHQ/hackwrld-reset/datastore"
	"github.com/DggHQ/hackwrld-reset/k8s"
	"github.com/gorilla/websocket"
)

type Msg struct {
	Data string `json:"data"`
}

var (
	etcdEndpoints = getEnvToArray("ETCD_ENDPOINTS", "10.10.90.5:2379;10.10.90.6:2379")
	namespace     = getEnv("NAMESPACE", "hackwrld")
	labelSelector = getEnv("LABEL_SELECTOR", "hackwrld-component=client")
	restartTime   = time.Now().Add(time.Minute * 30)
	u             = url.URL{
		Scheme:   getEnv("SCHEME", "ws"),
		Host:     fmt.Sprintf("%s:%s", getEnv("HOST", "localhost"), getEnv("PORT", "8080")),
		Path:     "/ws",
		RawQuery: fmt.Sprintf("token=%s", getEnv("KEY", "secret")),
	}
	wsmessage     = make(chan []byte)
	minutes       = 30
	operation     = getEnv("OPERATION", "DEPLOYMENTS")
	webDeployment = getEnv("WEB_DEPLOYMENT_NAME", "web")
)

// Handle setting of variables of env var is not set
func getEnvToArray(key, defaultValue string) []string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return strings.Split(defaultValue, ";")
	}
	return strings.Split(value, ";")
}

// Handle setting of variables of env var is not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func deleteDeployments() {
	// Load K8sConfig
	k8s := k8s.KubeManager{}
	ctx := context.TODO()
	k8s.Init().LoadClientSet()
	log.Println("Update Web Deployment to set maintenance to enabled")
	k8s.UpdateWebDeploymentEnv(ctx, namespace, webDeployment, "enabled")
	time.Sleep(time.Minute * 1)
	log.Println("Deleting Deployments")
	err := k8s.DeletePlayers(ctx, namespace, labelSelector)
	if err != nil {
		log.Println(err)
	}
	log.Println("Deleting deployments complete.")
	os.Exit(0)
}

func resetState() {
	log.Println("Deleting states.")
	// Reset the state of the whole game
	datastore := datastore.DataStore{}
	err := datastore.Init(etcdEndpoints, time.Second*5).ResetGame()
	if err != nil {
		log.Println(err)
	}
	log.Println("Deleting states complete.")
	time.Sleep(time.Minute * 1)
	k8s := k8s.KubeManager{}
	ctx := context.TODO()
	k8s.Init().LoadClientSet()
	log.Println("Update Web Deployment to set maintenance to disabled")
	k8s.UpdateWebDeploymentEnv(ctx, namespace, webDeployment, "disabled")

	os.Exit(0)
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
		// Launch goroutine that handles the messages that come into the channel and write them to the connection.
		go func(connection *websocket.Conn) {
			for msg := range wsmessage {
				// Dunnot if writedeadline is needed since it seems to be working fine w/o it
				//c.SetWriteDeadline(time.Now().Add(writeWait))
				ok := c.WriteMessage(websocket.TextMessage, msg)
				if ok != nil {
					log.Println("write:", ok)
				}
			}
		}(c)

		for {
			if time.Now().Before(restartTime) {
				log.Print("Waiting for restart. Letting players know")
				msg := Msg{
					Data: fmt.Sprintf("[HACKWRLD will start a new game in %d minutes]", minutes),
				}
				minutes = minutes - 10
				message, err := json.Marshal(msg)
				if err != nil {
					log.Fatalln(err)
				}
				// Write message to channel to be written to websocket connection
				wsmessage <- message
				// Sleep
				time.Sleep(time.Minute * 10)
				continue
			}
			log.Println("Timer done. Deleting current game")
			// Let players know game restarts
			msg := Msg{
				Data: "[HACKWRLD will now start a new game. Please login again to start your command center]",
			}
			message, err := json.Marshal(msg)
			if err != nil {
				log.Fatalln(err)
			}
			wsmessage <- message
			// Delete Deployments and exit
			deleteDeployments()
		}
	case "STATE":
		// Delete State
		resetState()
	}

}
