package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	// INITIAL gateway url
	gatewayUrl = "wss://gateway.discord.gg?v=10&encoding=json"

	// Internal makes
	websocketWrites = make(chan json.RawMessage)
	exit            = make(chan int)
	eventCallbacks  = make(map[int]func(data Event))

	// Internal variables
	wsconn         *websocket.Conn = nil
	debugMode                      = true
	nextCallbackId                 = 0
	callbackMutex  sync.RWMutex

	// Reconnecting
	lastSerial    *int       = nil
	lastSessionId *Snowflake = nil
	lastResumeUrl *string    = nil

	// Bot settings
	intents        = 33281
	token   string = ""
)

func debug(msg string) {
	if !debugMode {
		return
	}

	fmt.Print(msg + "\n")
}

func toJSON(v any) *[]byte {
	data, _ := json.Marshal(v)
	return &data
}

func init() {
	godotenv.Load()

	token = os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN not specified in environment variables. Critical.")
	}

	conn, _, err := websocket.DefaultDialer.Dial(gatewayUrl, nil)
	if err != nil {
		log.Fatalf("Error when initalizing websocket: %s", err.Error())
	}

	debug("Initalized connection to " + gatewayUrl)
	wsconn = conn
}

func websocketWriterLoop() {
	for {
		write := <-websocketWrites
		err := wsconn.WriteJSON(write)

		if err != nil {
			log.Printf("Error when writing to websocket: %s. Ignoring.", err.Error())
			continue
		}
	}
}

func onDisconnect(errorCode int) {
	reconnectCodes := []int{4000, 4001, 4002, 4003, 4005, 4007, 4008, 4009}
	shouldReconnect := slices.Contains(reconnectCodes, errorCode)

	if !shouldReconnect {
		log.Fatalf("Received non-recconectable exit code from discord: %d\n", errorCode)
	}

	conn, _, err := websocket.DefaultDialer.Dial(*lastResumeUrl, nil)
	if err != nil {
		log.Fatalf("Failed to reconnect to Discord: %s", err)
	}
	wsconn = conn
	debug("Reconnected to Discord")

	reconnect()
}

func websocketReaderLoop() {
	for {
		var event Event
		err := wsconn.ReadJSON(&event)

		if err != nil {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				debug(fmt.Sprintf("WebSocket closed by discord with code %d: %s\n", closeErr.Code, closeErr.Text))
				onDisconnect(closeErr.Code)
			} else {
				log.Fatalf("Error when reading to websocket (likely not discord related): %s. Critical.", err.Error())
			}
		}

		callbackMutex.RLock()
		for _, callback := range eventCallbacks {
			go callback(event)
		}
		callbackMutex.RUnlock()
	}
}

func addEventCallback(callback func(data Event)) func() {
	callbackMutex.Lock()
	thisCallbackId := nextCallbackId

	eventCallbacks[thisCallbackId] = callback
	nextCallbackId++

	callbackMutex.Unlock()

	return func() {
		callbackMutex.Lock()
		delete(eventCallbacks, thisCallbackId)
		callbackMutex.Unlock()
	}
}

func waitOpcode(opcode int) Event {
	waiter := make(chan bool)
	var eventData Event

	cancel := addEventCallback(func(data Event) {
		if data.Opcode == opcode {
			eventData = data
			waiter <- true
		}
	})

	<-waiter
	cancel()
	return eventData
}

func waitEventName(eventName string) Event {
	waiter := make(chan bool)
	var eventData Event

	cancel := addEventCallback(func(data Event) {
		if data.EventName == nil {
			return
		}

		if strings.EqualFold(eventName, *data.EventName) {
			waiter <- true
			eventData = data
		}
	})

	<-waiter
	cancel()
	return eventData
}

func waitReady() {
	waitEventName("READY")
}

func writeToWebsocket(data []byte) {
	websocketWrites <- data
}

func main() {
	debug("Hello, world!")

	go websocketReaderLoop()
	go websocketWriterLoop()

	if debugMode {
		addEventCallback(func(data Event) {
			switch data.Opcode {
			case HeartbeatAcknowledgeOpcode:
				debug("Discord has acknowledged our heartbeat")
			case ReadyOpcode:
				if data.EventName == nil || *data.EventName != "READY" {
					return
				}

				debug("Ready event received.")
			default:
				debug("Received unknown opcode - " + strconv.Itoa(data.Opcode) + ". Raw data: \n" + string(data.Data))
			}

			if data.Serial != nil {
				debug(fmt.Sprintf("(Serial: %d)", *data.Serial))
			}
		})
	}

	<-exit
}
