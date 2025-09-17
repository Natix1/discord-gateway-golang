package main

import (
	"encoding/json"
	"log"
)

func identify(data IdentifyData) {
	serialized, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Failed while serializing Identify event data: %s", err.Error())
	}

	writeToSocket(serialized)
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
