package main

import (
	"encoding/json"
	"log"
)

func init() {
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
}
