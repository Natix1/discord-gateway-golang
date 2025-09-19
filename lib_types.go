package discord

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type CallbackID = int
type CallbackFunction func(event *Event)
type Snowflake = string
type Opcode = int
type EventName = string

type BotClient struct {
	token            string
	context          context.Context
	intents          int
	ready            bool
	heartbeatRunning bool

	callbacks      map[int]*Callback
	callbackMutex  sync.RWMutex
	nextCallbackID int

	lastSessionID     *string
	lastReconnectURL  *string
	lastSerial        *int
	heartbeatInterval *time.Duration

	websocketConnection *websocket.Conn
	websocketWriteQueue chan json.RawMessage
}

type Callback struct {
	id           int
	callbackType int
	client       *BotClient
	function     *CallbackFunction
}
