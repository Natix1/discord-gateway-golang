package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

var (
	mainOnce sync.Once
)

func startMain() {
	mainOnce.Do(func() {
		go main()
	})
}

func TestMessages(t *testing.T) {
	startMain()
	waitReady()

	channel := PartialChannel{
		ID: Snowflake("1327997437805334539"),
	}

	// NONCE MESSAGE

	confirmedWait := make(chan bool)
	nonce := channel.SendDeferred("[Unit test] I am a deferred message.")

	cancel := addEventCallback(func(data Event) {
		if !(data.EventName != nil && *data.EventName == "MESSAGE_CREATE") {
			return
		}

		var message Message
		err := json.Unmarshal(data.Data, &message)
		if err != nil {
			t.Fatal(err.Error())
		}

		if message.Nonce == nil {
			return
		}

		if *message.Nonce == nonce {
			confirmedWait <- true
			fmt.Printf("Nonce worked. EUREKA. Nonce = %s", nonce)
		}
	})

	<-confirmedWait
	cancel()

	// REGULAR MESSAGE

	_, err := channel.Send("[Unit test] Hello, world! I, on the other hand, am a normal message")
	if err != nil {
		t.Fatal(err.Error())
	}
}
