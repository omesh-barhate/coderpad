package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/omesh-barhate/coderpad/commons"
)

type ClientInfo struct {
	Username string `json:"username"`
	SiteID   string `json:"siteID"`
	Conn     *websocket.Conn
}

var (
	nextSiteID     = 0
	siteIDMutex    sync.Mutex
	wsUpgrader     = websocket.Upgrader{}
	activeClients  = make(map[uuid.UUID]ClientInfo)
	messageChannel = make(chan commons.Message)
	syncChannel    = make(chan commons.Message)
)

func main() {
	address := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleWebSocket)

	go syncHandler()
	go messageHandler()

	log.Printf("Starting server on %s", *address)

	server := &http.Server{
		Addr:         *address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error starting server, exiting.", err)
	}
}

func handleWebSocket(response http.ResponseWriter, request *http.Request) {
	clientConnection, err := wsUpgrader.Upgrade(response, request, nil)
	if err != nil {
		color.Red("WebSocket upgrade error: %v\n", err)
	}
	defer clientConnection.Close()

	color.Yellow("active clients: %d\n", len(activeClients))

	clientID := uuid.New()

	siteIDMutex.Lock()
	nextSiteID++
	siteIDMutex.Unlock()

	siteIDString := strconv.Itoa(nextSiteID)

	clientInfo := ClientInfo{Conn: clientConnection, SiteID: siteIDString}
	activeClients[clientID] = clientInfo

	color.Magenta("clients after SiteID: %+v", activeClients)
	color.Yellow("Assigned siteID: %s", clientInfo.SiteID)

	siteIDMessage := commons.Message{MessageType: commons.SiteIDMessage, Text: clientInfo.SiteID, ClientID: clientID}
	if err := clientConnection.WriteJSON(siteIDMessage); err != nil {
		color.Red("Failed to send siteID message")
	}

	for id, info := range activeClients {
		if id != clientID {
			message := commons.Message{MessageType: commons.DocReqMessage, ClientID: clientID}
			color.Cyan("sending docReq to %s for %s", id, clientID)
			if err := info.Conn.WriteJSON(&message); err != nil {
				color.Red("Failed to send docReq: %v\n", err)
				continue
			}
			break
		}
	}

	for {
		var message commons.Message
		if err := clientConnection.ReadJSON(&message); err != nil {
			color.Red("Closing connection for username: %v\n", activeClients[clientID].Username)
			delete(activeClients, clientID)
			break
		}
		message.ClientID = clientID
		if message.MessageType == commons.DocSyncMessage {
			syncChannel <- message
			continue
		}
		messageChannel <- message
	}
}

func messageHandler() {
	for {
		message := <-messageChannel
		timestamp := time.Now().Format(time.ANSIC)
		if message.MessageType == commons.JoinMessage {
			info := activeClients[message.ClientID]
			info.Username = message.Username
			activeClients[message.ClientID] = info
			color.Green("%s >> %s %s (ID: %s)\n", timestamp, message.Username, message.Text, message.ClientID)
		} else if message.MessageType == "operation" {
			color.Green("operation >> %+v from ID=%s\n", message.Operation, message.ClientID)
		} else {
			color.Green("%s >> %+v\n", timestamp, message)
		}
		for id, info := range activeClients {
			if id != message.ClientID {
				color.Magenta("writing message to: %s, msg: %+v\n", id, message)
				if err := info.Conn.WriteJSON(message); err != nil {
					color.Red("Send error: %v\n", err)
					info.Conn.Close()
					delete(activeClients, id)
				}
			}
		}
	}
}

func syncHandler() {
	for {
		syncMessage := <-syncChannel
		color.Cyan("got syncMsg, len(document) = %d\n", len(syncMessage.Document.Characters))
		for id, info := range activeClients {
			if id != syncMessage.ClientID {
				color.Cyan("sending syncMsg to %s", syncMessage.ClientID)
				_ = info.Conn.WriteJSON(syncMessage)
			}
		}
	}
}
