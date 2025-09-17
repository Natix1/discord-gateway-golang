package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	gatewayUrl = "wss://gateway.discord.gg?v=10&encoding=json"
)

var (
	webhookWrites  = make(chan json.RawMessage)
	eventCallbacks = []func(data Event){}

	wsconn       *websocket.Conn = nil
	debugMode                    = true
	debugVerbose                 = true

	token     string = ""
	topSerial *int   = nil
)

func debug(msg string) {
	if !debugMode {
		return
	}

	fmt.Print(msg + "\n")
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

	go websocketReaderLoop()
	go websocketWriterLoop()
}

func websocketWriterLoop() {
	for {
		write := <-webhookWrites
		err := wsconn.WriteJSON(write)
		if debugVerbose {
			debug(fmt.Sprintf("Wrote event to discord. Raw data: \n%v", string(write)))
		} else {
			debug("Wrote event to discord.")
		}

		if err != nil {
			log.Printf("Error when writing to websocket: %s. Ignoring.", err.Error())
			continue
		}
	}
}

func websocketReaderLoop() {
	for {
		var event Event
		err := wsconn.ReadJSON(&event)

		if err != nil {
			log.Printf("Error when reading to websocket: %s. Ignoring.", err.Error())
			continue
		}

		if debugVerbose {
			debug(fmt.Sprintf("Got event from discord. Opcode: %d. Raw data: \n%v", event.Opcode, string(event.Data)))
		} else {
			debug("Got event from discord")
		}

		for _, callback := range eventCallbacks {
			go callback(event)
		}
	}
}

func addEventCallback(callback func(data Event)) {
	eventCallbacks = append(eventCallbacks, callback)
}

func writeToSocket(data []byte) {
	webhookWrites <- data
}

func main() {
	addEventCallback(func(rawData Event) {
		if rawData.Opcode != HelloOpcode {
			return
		}

		var data HelloData
		err := json.Unmarshal(rawData.Data, &data)
		if err != nil {
			log.Fatalf("Failed decoding HelloOpcode event data: %s. Critical.", err.Error())
		}

		debug("Received Hello event, starting heartbeat loop...")
		go startHeartbeatLoop(data.HeartbeatInterval)
	})

	addEventCallback(func(data Event) {
		if data.Serial == nil {
			return
		}

		if topSerial == nil {
			topSerial = data.Serial
		} else {
			if *topSerial < *data.Serial {
				topSerial = data.Serial
			}
		}
	})

	if debugMode {
		addEventCallback(func(data Event) {
			switch data.Opcode {
			case HeartbeatAcknowledgeOpcode:
				debug("Discord has acknowledged our heartbeat")
			default:
				debug("Received unknown opcode: " + strconv.Itoa(data.Opcode))
			}
		})
	}

	select {}
}
