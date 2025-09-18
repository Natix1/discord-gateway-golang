package discordgateway

import (
	"fmt"
	"math/rand"
	"time"
)

func getJitter() float64 {
	r := rand.Float64()
	return r
}

func sendHeartbeat() {
	data := struct {
		Opcode int  `json:"op"`
		Serial *int `json:"d"`
	}{
		Opcode: 1,
		Serial: lastSerial,
	}

	debug("Sending heartbeat...")
	writeToWebsocket(*toJSON(data))
}

func startHeartbeatLoop(initial_interval int) {
	waitTime := initial_interval

	for {
		jitter := getJitter()
		waitJitter := time.Duration(float64(waitTime)*jitter) * time.Millisecond

		debug(fmt.Sprintf("Waiting %dms before sending our heartbeat... (jitter = %.3f, waitTime = %dms, total wait time = %dms)",
			waitJitter.Milliseconds(), jitter, waitTime, waitJitter.Milliseconds()))

		time.Sleep(waitJitter)
		sendHeartbeat()
	}
}
