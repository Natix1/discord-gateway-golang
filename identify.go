package main

import (
	"encoding/json"
	"log"
)

const (
	intents = 32769
)

func init() {
	addEventCallback(func(data Event) {
		if data.Opcode != HelloOpcode {
			return
		}

		identifyData := IdentifyData{
			Token:   token,
			Intents: intents,
			Properties: struct {
				OperatingSystem string `json:"os"`
				Browser         string `json:"browser"`
				Device          string `json:"device"`
			}{
				OperatingSystem: "Linux",
				Browser:         "github.com/natix1/discord-gateway-golang",
				Device:          "github.com/natix1/discord-gateway-golang",
			},
		}

		identifyDataSerialized, err := json.Marshal(identifyData)
		if err != nil {
			log.Fatalf("Failed serializing identify data: %s", err.Error())
		}

		eventData := Event{
			Opcode: IdentifyOpcode,
			Data:   identifyDataSerialized,
		}

		eventDataSerialized, err := json.Marshal(eventData)
		if err != nil {
			log.Fatalf("Failed serializing event data: %s", err.Error())
		}

		writeToWebsocket(eventDataSerialized)
	})
}

/*
identify(IdentifyData{
	Token:   token,
	Intents: 513,
	Properties: struct {
		OperatingSystem string "json:\"os\""
		Browser         string "json:\"browser\""
		Device          string "json:\"device\""
	}{
		OperatingSystem: "Linux",
		Browser:         "github.com/natix1/discord-gateway-golang",
		Device:          "github.com/natix1/discord-gateway-golang",
	},
})
*/
