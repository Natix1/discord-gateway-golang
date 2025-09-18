package main

import (
	"encoding/json"
	"log"
)

func reconnect() {
	if lastResumeUrl == nil || lastSessionId == nil {
		log.Fatal("Can't reconnect. Resume URL or Session ID invalid.")
	}

	resumeData := ResumeData{
		Token:      token,
		SessionID:  *lastSessionId,
		LastSerial: *lastSerial,
	}

	eventData := Event{
		Opcode: ResumeOpcode,
		Data:   *toJSON(resumeData),
	}

	writeToWebsocket(*toJSON(eventData))
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
