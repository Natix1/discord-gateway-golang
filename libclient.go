package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultURL         = "wss://gateway.discord.gg?v=10&encoding=json"
	version            = "v0.0.1"
	clientName         = "github.com/natix1/discord-gateway-golang"
	debugMode          = true
	debugVerboseOutput = false
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

func toJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

func (client *BotClient) addCallback(eventType int, callback CallbackFunction) *Callback {
	client.callbackMutex.Lock()
	defer client.callbackMutex.Unlock()

	client.nextCallbackID++
	cb := &Callback{
		id:           client.nextCallbackID,
		callbackType: eventType,
		client:       client,
		function:     &callback,
	}

	client.callbacks[client.nextCallbackID] = cb
	return cb
}

func (client *BotClient) OnEvent(eventName string, callback CallbackFunction) *Callback {
	eventName = strings.ToLower(eventName)

	return client.addCallback(_EventNameCallbackType, func(event *Event) {
		if event.EventName != nil && strings.EqualFold(*event.EventName, eventName) {
			go callback(event)
		}
	})
}

func (client *BotClient) OnReady(callback func()) *Callback {
	if client.ready {
		go callback()
		return nil
	} else {
		return client.addCallback(_ReadyCallbackType, func(_ *Event) {
			go callback()
		})
	}
}

func (client *BotClient) OnOpcode(opcode int, callback CallbackFunction) *Callback {
	return client.addCallback(_OpcodeCallbackType, func(event *Event) {
		if event.Opcode == opcode {
			go callback(event)
		}
	})
}

func (client *BotClient) OnAnyEvent(callback func(event *Event)) *Callback {
	return client.addCallback(_AnyCallbackType, callback)
}

func (client *BotClient) DebugPrint(message string) {
	if !debugMode {
		return
	}

	log.Print(message + "\n")
}

func (client *BotClient) onWebsocketError(err error) {
	if _, ok := err.(*net.OpError); ok {
		return
	}

	var closeErr *websocket.CloseError
	code := 0
	if errors.As(err, &closeErr) {
		code = closeErr.Code
	}

	reconnectCodes := []int{4000, 4001, 4002, 4003, 4005, 4007, 4008, 4009}
	if slices.Contains(reconnectCodes, code) || code == 0 {
		client.DebugPrintf("Websocket error, attempting reconnect: %v", err)
		client.reconnect()
		return
	}

	log.Fatal(err)
}

func (client *BotClient) DebugPrintf(message string, args ...any) {
	if !debugMode {
		return
	}

	res := fmt.Sprintf(message, args...)
	log.Print(res + "\n")
}

func (client *BotClient) GetFullClientName() string {
	return fmt.Sprintf("%s (%s, %s)", clientName, clientName, version)
}

// Cleans everything up, closes any open websocket, frees things
func (client *BotClient) Cleanup() {
	client.DebugPrint("Cleanup() triggered")

	if client.websocketConnection != nil {
		message := websocket.FormatCloseMessage(1000, "Cleanup")
		err := client.websocketConnection.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		if err != nil && debugMode {
			client.DebugPrint("Failed disposing of websocket gracefully.")
		}

		client.websocketConnection.Close()
		client.websocketConnection = nil
	}

	for _, cb := range client.callbacks {
		cb.Disconnect()
	}
}

func (client *BotClient) bindWebsocket() {
	// Read
	go func() {
		for {
			if client.websocketConnection == nil {
				break
			}

			var event Event
			err := client.websocketConnection.ReadJSON(&event)
			if err != nil {
				client.onWebsocketError(err)
			}

			client.callbackMutex.RLock()

			for _, handler := range client.callbacks {
				cbFunction := *handler.function

				switch handler.callbackType {
				case _OpcodeCallbackType:
					if event.Opcode <= 0 {
						break
					}

					go cbFunction(&event)
				case _EventNameCallbackType:
					if event.EventName != nil {
						go cbFunction(&event)
					}
				case _AnyCallbackType:
					go cbFunction(&event)
				case _ReadyCallbackType:
					if event.EventName == nil {
						break
					}

					if strings.EqualFold(*event.EventName, "ready") {
						go cbFunction(&event)
					}
				}
			}

			client.callbackMutex.RUnlock()
		}
	}()

	// Write
	go func() {
		for {
			thisWrite := <-client.websocketWriteQueue

			if client.websocketConnection == nil {
				break
			}

			if debugMode && debugVerboseOutput {
				lines := ""
				lines += "=================== WEBSOCKET WRITE ===================\n"
				lines += fmt.Sprintf("Raw bytes: \n%v\n", string(thisWrite))
				client.DebugPrint(lines)
			}

			err := client.websocketConnection.WriteMessage(websocket.TextMessage, thisWrite)
			if err != nil {
				client.onWebsocketError(err)
			}
		}
	}()
}

func (client *BotClient) startHeartbeatLoop() {
	if client.heartbeatRunning {
		return
	}

	client.heartbeatRunning = true

	sendHeartbeat := func() {
		data := HeartbeatEvent{
			Opcode:     HeartbeatOpcode,
			LastSerial: client.lastSerial,
		}

		client.DebugPrint("Sending heartbeat...")
		client.writeSocket(toJSON(data))
	}

	jitter := rand.Float64()
	time.Sleep(time.Millisecond * time.Duration(float64(client.heartbeatInterval.Milliseconds())*jitter))
	sendHeartbeat()

	ticker := time.NewTicker(*client.heartbeatInterval)
	go func() {
		for {
			if client.websocketConnection == nil {
				break
			}

			<-ticker.C

			if client.websocketConnection == nil {
				break
			}

			sendHeartbeat()
		}

		ticker.Stop()
		client.heartbeatRunning = false
	}()
}

func (client *BotClient) writeSocket(data json.RawMessage) {
	client.websocketWriteQueue <- data
}

func (client *BotClient) identify() {
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
		Data:   toJSON(identifyData),
	}

	client.writeSocket(toJSON(eventData))
}

func (client *BotClient) reconnect() error {
	if client.websocketConnection == nil {
		client.Run()
	}

	if client.lastSessionID == nil || client.lastSerial == nil {
		return errors.New("invalid session. lastSerial and/or lastSessionID are nil.")
	}

	resumeData := ResumeData{
		Token:      client.token,
		SessionID:  *client.lastSessionID,
		LastSerial: *client.lastSerial,
	}

	event := Event{
		Opcode: ResumeOpcode,
		Data:   toJSON(resumeData),
	}

	client.writeSocket(toJSON(event))
	return nil
}

func (client *BotClient) bindDefaultEvents() {
	client.OnOpcode(HelloOpcode, func(event *Event) {
		var data HelloData
		err := json.Unmarshal(event.Data, &data)
		if err != nil {
			log.Fatal(err)
		}

		interval := time.Duration(data.HeartbeatInterval) * time.Millisecond
		client.heartbeatInterval = &interval
		client.DebugPrintf("Got hello event (opcode %d) event from discord. Heartbeat interval = %dms. Sending identify opcode.", HelloOpcode, interval.Milliseconds())
		go client.startHeartbeatLoop()
		go client.identify()
	})

	client.OnAnyEvent(func(event *Event) {
		if debugMode && debugVerboseOutput {
			lines := ""
			lines += "==================== EVENT RECEIVED ====================\n"
			lines += fmt.Sprintf("Opcode %d\n", event.Opcode)
			if event.EventName == nil {
				lines += "Event name: -\n"
			} else {
				lines += fmt.Sprintf("Event name: %v\n", *event.EventName)
			}

			lines += fmt.Sprintf("Raw data: \n%v\n", string(event.Data))
			client.DebugPrint(lines)
		}
		if event.Serial == nil {
			return
		}

		if client.lastSerial == nil {
			client.lastSerial = event.Serial
			client.DebugPrint("First serial: " + strconv.Itoa(*event.Serial))
		} else {
			if *client.lastSerial < *event.Serial {
				client.lastSerial = event.Serial
				client.DebugPrint("New highest serial: " + strconv.Itoa(*event.Serial))
			}
		}
	})

	client.OnOpcode(ReconnectOpcode, func(event *Event) {
		client.reconnect()
	})

	client.OnOpcode(InvalidSessionOpcode, func(event *Event) {
		var b bool
		if err := json.Unmarshal(event.Data, &b); err == nil {
			if b {
				client.reconnect()
			} else {
				client.Cleanup()
				client.Run()
			}
		}
	})

	client.addCallback(_ReadyCallbackType, func(_ *Event) {
		client.ready = true
	})

	if debugMode {
		client.OnOpcode(HeartbeatAcknowledgeOpcode, func(event *Event) {
			client.DebugPrint("Heartbeat acknowledged.")
		})
	}
}

func (client *BotClient) Run() error {
	client.Cleanup()

	url := defaultURL
	if client.lastReconnectURL != nil {
		url = *client.lastReconnectURL
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	client.websocketConnection = conn
	client.bindWebsocket()
	client.bindDefaultEvents()

	return nil
}

func New(token string, intents int, context context.Context) BotClient {
	return BotClient{
		token:   token,
		context: context,
		intents: intents,
		ready:   false,

		callbacks:      make(map[int]*Callback),
		callbackMutex:  sync.RWMutex{},
		nextCallbackID: 0,

		lastSessionID:     nil,
		lastReconnectURL:  nil,
		lastSerial:        nil,
		heartbeatInterval: nil,

		websocketWriteQueue: make(chan json.RawMessage),
	}
}
