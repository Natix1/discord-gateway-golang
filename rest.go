package discord

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	MAX_MESSAGE_LEN = 2000
)

const (
	METHOD_GET   = "GET"
	METHOD_POST  = "POST"
	METHOD_FETCH = "FETCH"
	METHOD_PATCH = "PATCH"
)

var (
	BASE_URL    = "https://discord.com/api/v10"
	HTTP_CLIENT = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func (client *BotClient) restRequest(method string, path string, data []byte) ([]byte, error) {
	if len([]rune(path)) > MAX_MESSAGE_LEN {
		return []byte{}, errors.New("message exceeds max message lenght")
	}

	client.DebugPrintf("Sending %s request to %s%s", method, BASE_URL, path)

	requestUrl := BASE_URL + path

	var reader io.Reader = nil
	if data != nil {
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, requestUrl, reader)
	if err != nil {
		log.Printf("[%s] [%s] - FAILED: %s", method, requestUrl, err.Error())
		return []byte{}, err
	}

	req.Header.Set("Authorization", "Bot "+client.token)
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
