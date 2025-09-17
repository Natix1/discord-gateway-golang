package main

import "encoding/json"

const (
	HelloOpcode                = 10
	HeartbeatOpcode            = 1
	HeartbeatAcknowledgeOpcode = 11
	IdentifyOpcode             = 2
)

type Event struct {
	Opcode    int             `json:"op"`
	Data      json.RawMessage `json:"d"`
	EventName *string         `json:"t"`
	Serial    *int            `json:"s"`
}

type IdentifyData struct {
	Token      string `json:"token"`
	Intents    int    `json:"intents"`
	Properties struct {
		OperatingSystem string `json:"os"`
		Browser         string `json:"browser"`
		Device          string `json:"device"`
	} `json:"properties"`
}

type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type HeartbeatAcknowledge struct {
}
