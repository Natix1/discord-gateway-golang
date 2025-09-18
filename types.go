package main

import "encoding/json"

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

type ReadyData struct {
	ResumeGatewayURL string `json:"resume_gateway_url"`
	SessionID        string `json:"session_id"`
}

type ResumeData struct {
	Token      string `json:"token"`
	SessionID  string `json:"session_id"`
	LastSerial int    `json:"seq"`
}
