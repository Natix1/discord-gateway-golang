package discordgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultURL = "wss://gateway.discord.gg?v=10&encoding=json"
	version    = "v0.0.1"
	clientName = "github.com/natix1/discord-gateway-golang"
)

const (
	HelloOpcode                = 10
	HeartbeatOpcode            = 1
	HeartbeatAcknowledgeOpcode = 11
	IdentifyOpcode             = 2
	ReadyOpcode                = 0
	ReconnectOpcode            = 7
	InvalidSessionOpcode       = 9
	ResumeOpcode               = 6
)

type CallbackType interface {
	~int | ~string
}

func toJSON(v any) *[]byte {
	data, _ := json.Marshal(v)
	return &data
}

func (client *LibClient) OnEvent(eventName string, callback func(event *Event)) EventDispatcher {
	eventName = strings.ToLower(eventName)

	client.callbackMutex.Lock()
	defer client.callbackMutex.Unlock()

	client.nextCallbackID++
	client.eventCallbacks[eventName][client.nextCallbackID] = callback

	return EventDispatcher{
		eventNameOrOpcode: eventName,
		eventType:         1,
		id:                client.nextCallbackID,
		client:            client,
	}
}

func (client *LibClient) OnReady(callback func()) {
	if client.ready {
		callback()
	} else {
		client.readyCallbacks = append(client.readyCallbacks, callback)
	}
}

func (client *LibClient) OnOpcode(opcode int, callback func(event *Event)) EventDispatcher {
	client.callbackMutex.Lock()
	defer client.callbackMutex.Unlock()

	client.nextCallbackID++
	client.opcodeCallbacks[opcode][client.nextCallbackID] = callback

	return EventDispatcher{
		eventNameOrOpcode: opcode,
		eventType:         0,
		id:                client.nextCallbackID,
		client:            client,
	}
}

func (client *LibClient) OnAnyEvent(callback func(event *Event)) EventDispatcher {
	client.callbackMutex.Lock()
	defer client.callbackMutex.Unlock()
	client.nextCallbackID++
	client.anyEventCalblacks = append(client.anyEventCalblacks, callback)

	return EventDispatcher{
		eventNameOrOpcode: "any",
		eventType:         2,
		id:                client.nextCallbackID,
		client:            client,
	}
}

func (client *LibClient) DebugPrint(message string) {
	if !client.debugMode {
		return
	}

	log.Print(message + "\n")
}

func (client *LibClient) DebugPrintf(message string, args ...any) {
	if !client.debugMode {
		return
	}

	res := fmt.Sprintf(message, args...)
	log.Print(res + "\n")
}

func (client *LibClient) GetFullClientName() string {
	return fmt.Sprintf("%s (%s, %s)", clientName, clientName, version)
}

// Cleans everything up, closes any open websocket, frees things
func (client *LibClient) Cleanup() {
	client.DebugPrint("Cleanup() triggered")

	if client.websocketConnection != nil {
		message := websocket.FormatCloseMessage(1000, "Cleanup")
		err := client.websocketConnection.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		if err != nil && client.debugMode {
			client.DebugPrint("Failed disposing of websocket gracefully.")
		}

		client.websocketConnection.Close()
		client.websocketConnection = nil
	}
}

func (client *LibClient) bindWebsocket() {
	// Read
	go func() {
		for {
			if client.websocketConnection == nil {
				break
			}

			var event Event
			err := client.websocketConnection.ReadJSON(&event)
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok || client.websocketConnection == nil {
					return
				}

				log.Fatalf("Websocket write caused unknown exception: %s\n", err.Error())
			}

			opcodeEvents := client.opcodeCallbacks[event.Opcode]

			client.callbackMutex.RLock()
			defer client.callbackMutex.RUnlock()

			for _, handler := range opcodeEvents {
				go handler(&event)
			}

			if event.EventName != nil {
				nameEvents := client.eventCallbacks[strings.ToLower(*event.EventName)]

				for _, handler := range nameEvents {
					go handler(&event)
				}
			}
		}
	}()

	// Write
	go func() {
		for {
			thisWrite := <-client.websocketWriteQueue

			if client.websocketConnection == nil {
				break
			}

			err := client.websocketConnection.WriteMessage(websocket.TextMessage, thisWrite)
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok || client.websocketConnection == nil {
					return
				}

				log.Fatalf("Websocket write caused unknown exception: %s\n", err.Error())
			}
		}
	}()
}

func (client *LibClient) startHeartbeatLoop() {
	go func() {
		for {
			if client.websocketConnection == nil {
				time.Sleep(time.Second)
				continue
			}

			if client.heartbeatInterval == nil {
				time.Sleep(time.Second)
				continue
			}

			jitter := rand.Float32()
			waitTimeMs := client.heartbeatInterval.Milliseconds() * int64(jitter)
			waitTime := time.Duration(waitTimeMs) * time.Millisecond

			client.DebugPrintf(
				"Waiting for next heartbeat... (jitter = %d, interval = %d, final = %d)",
				waitTimeMs,
				client.heartbeatInterval.Milliseconds(),
				waitTimeMs,
			)

			time.Sleep(waitTime)

			data := HeartbeatEvent{
				Opcode:     HeartbeatOpcode,
				LastSerial: client.lastSerial,
			}

			client.writeSocket(*toJSON(data))
		}
	}()
}

func (client *LibClient) writeSocket(data []byte) {
	client.websocketWriteQueue <- data
}

func (client *LibClient) identify() {
	identifyData := IdentifyData{
		Token:   client.token,
		Intents: client.intents,
		Properties: IdentifyProperties{
			OperatingSystem: "?",
			Browser:         client.GetFullClientName(),
			Device:          client.GetFullClientName(),
		},
	}

	eventData := Event{
		Opcode: IdentifyOpcode,
		Data:   *toJSON(identifyData),
	}

	client.writeSocket(*toJSON(eventData))
}

func (client *LibClient) bindDefaultEvents() {
	client.OnOpcode(HelloOpcode, func(event *Event) {
		var data HelloData
		err := json.Unmarshal(event.Data, &data)
		if err != nil {
			log.Fatal(err)
		}

		interval := time.Duration(data.HeartbeatInterval) * time.Millisecond
		client.heartbeatInterval = &interval
		client.DebugPrintf("Got hello event (opcode %d) event from discord. Heartbeat interval = %dms.", HelloOpcode, interval.Milliseconds())
	})

	client.OnEvent("ready", func(event *Event) {
		var readyData ReadyData
		err := json.Unmarshal(event.Data, &readyData)
		if err != nil {
			log.Fatal(err)
		}

		client.lastReconnectURL = &readyData.ResumeGatewayURL
		client.lastSessionID = &readyData.SessionID

		client.callbackMutex.RLock()
		defer client.callbackMutex.RUnlock()
		for _, handler := range client.readyCallbacks {
			go handler()
		}

		client.readyCallbacks = client.readyCallbacks[:0]
		client.ready = true
	})

	client.OnAnyEvent(func(event *Event) {
		if event.Serial == nil {
			return
		}

		if client.lastSerial == nil {
			client.lastSerial = event.Serial
			client.DebugPrint("First serial: " + string(*event.Serial))
		} else {
			if *client.lastSerial < *event.Serial {
				client.lastSerial = event.Serial
				client.DebugPrint("New highest serial: " + string(*event.Serial))
			}
		}
	})

	if client.debugMode {
		client.OnOpcode(HeartbeatAcknowledgeOpcode, func(event *Event) {
			client.DebugPrint("Heartbeat acknowledged.")
		})
	}
}

// Initalizes the websocket connection and returns a cleanup function to call after you're done
func (client *LibClient) Run() error {
	client.Cleanup()

	conn, _, err := websocket.DefaultDialer.Dial(defaultURL, nil)
	if err != nil {
		return err
	}

	client.websocketConnection = conn
	client.bindWebsocket()
	client.bindDefaultEvents()

	return nil
}

func New(token string, intents int, context context.Context, debug bool) LibClient {
	return LibClient{
		token:   token,
		context: context,
		intents: intents,
		ready:   false,

		debugMode:         debug,
		eventCallbacks:    make(map[string][]func(data *Event)),
		opcodeCallbacks:   make(map[int][]func(data *Event)),
		readyCallbacks:    []func(){},
		anyEventCalblacks: []func(data *Event){},
		callbackMutex:     sync.RWMutex{},
		nextCallbackID:    0,

		lastSessionID:     nil,
		lastReconnectURL:  nil,
		lastSerial:        nil,
		heartbeatInterval: nil,

		websocketWriteQueue: make(chan []byte),
	}
}
