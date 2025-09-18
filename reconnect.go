package main

import (
	"encoding/json"
	"log"
)

func reconnect() {
	if lastResumeUrl == nil || lastSessionId == nil {
		log.Fatal("Can't reconnect. Resume URL or Session ID invalid.")
	}

	resumeData, err := json.Marshal(ResumeData{
		Token:      token,
		SessionID:  *lastSessionId,
		LastSerial: *lastSerial,
	})

	if err != nil {
		log.Fatal("Failed serializing resume data.")
	}

	eventData, err := json.Marshal(Event{
		Opcode: ResumeOpcode,
		Data:   resumeData,
	})

	if err != nil {
		log.Fatal("Failed serializing event data with resume data.")
	}

	writeToWebsocket(eventData)
}

func init() {
	addEventCallback(func(data Event) {
		if data.Opcode == ReconnectOpcode {
			reconnect()
		}

		if data.EventName != nil && *data.EventName == "READY" {
			var readyData ReadyData
			err := json.Unmarshal(data.Data, &readyData)
			if err != nil {
				log.Fatal("Failed decoding READY event. Critical.")
			}

			lastResumeUrl = &readyData.ResumeGatewayURL
			lastSessionId = &readyData.SessionID
		}
	})
}
