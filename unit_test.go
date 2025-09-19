package discord

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var bot *BotClient

func getToken() string {
	godotenv.Load()

	token := os.Getenv("TOKEN")
	if token == "" {
		panic("TOKEN not set")
	}

	return token
}

func TestMain(m *testing.M) {
	log.Print("Starting TestMain bot...")

	c := New(getToken(), 33281, context.TODO())
	bot = &c
	bot.Run()

	code := m.Run()

	bot.Cleanup()
	log.Print("Tests over...")
	os.Exit(code)
}

func TestLogin(t *testing.T) {
	waiter := make(chan bool)

	bot.OnReady(func() {
		t.Log("Logged in")
		waiter <- true
	})

	<-waiter
}

func TestMessage(t *testing.T) {
	bot.OnReady(func() {
		channel := bot.GetChannel("1327997437805334539")
		msg, err := channel.SendMessage("Hello, world! This is a unit test.")
		if err != nil {
			t.Fatal(err.Error())
		}

		t.Log("Message sent! ID: " + msg.ID)
	})
}
