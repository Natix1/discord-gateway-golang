package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	MAX_MESSAGE_LEN = 2000
)

var (
	BASE_URL    = "https://discord.com/api/v10"
	HTTP_CLIENT = &http.Client{
		Timeout: 10 * time.Second,
	}
	NONCES = make(map[string]bool)
)

func generateNonce() string {
	runeset := []rune("abcdefghijklmnopqrstuwxyz123456789!@#$%^&*-_=+{[()]};:'/.,<>?")
	nonce := []rune{}

	for range 25 {
		randRune := rand.Intn(len(runeset))
		nonce = append(nonce, runeset[randRune])
	}

	return string(nonce)
}

func request(method string, path string, data *[]byte) ([]byte, error) {
	if len([]rune(path)) > MAX_MESSAGE_LEN {
		return []byte{}, errors.New("message exceeds max message lenght")
	}

	debug(fmt.Sprintf("Sending %s request to %s%s", method, BASE_URL, path))

	requestUrl := BASE_URL + path

	var reader io.Reader = nil
	if data != nil {
		reader = bytes.NewReader(*data)
	}

	req, err := http.NewRequest(method, requestUrl, reader)
	if err != nil {
		log.Printf("[%s] [%s] - FAILED: %s", method, requestUrl, err.Error())
		return []byte{}, err
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("User-Agent", "discord-gateway-golang (https://github.com/natix1/discord-gateway-golang, test-v0.0.1)")
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := HTTP_CLIENT.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	allData, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []byte{}, errors.New("non-200 status. body: \n" + string(allData) + "\n")
	}

	return allData, nil
}

func sendMessage(channelID Snowflake, content string, reference *MessageReference, nonce *string) (*Message, error) {
	data := MessageSendData{
		Content:          &content,
		MessageReference: reference,
		Nonce:            nonce,
	}

	if nonce != nil {
		NONCES[*nonce] = true
	}

	response, err := request("POST", fmt.Sprintf("/channels/%s/messages", channelID), toJSON(data))
	if err != nil {
		return nil, err
	}

	var message Message
	err = json.Unmarshal(response, &message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (pc *PartialChannel) Send(content string) (*Message, error) {
	return sendMessage(pc.ID, content, nil, nil)
}

func (pc *PartialChannel) SendDeferred(content string) string {
	nonce := generateNonce()
	go sendMessage(pc.ID, content, nil, &nonce)
	return nonce
}

func (msg *Message) Reply(content string) (*Message, error) {
	return sendMessage(msg.ChannelID, content, &MessageReference{
		Type:      DefaultMessageReference,
		MessageID: &msg.ID,
		ChannelID: &msg.ChannelID,
		GuildID:   msg.GuildID,
	}, nil)
}
