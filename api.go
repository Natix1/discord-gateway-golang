package discord

import (
	"encoding/json"
	"fmt"
)

func (client *BotClient) makeObject(id Snowflake) Object {
	return Object{
		ID:        id,
		Available: false,
		client:    client,
	}
}

func (client *BotClient) GetChannel(id Snowflake) Channel {
	return Channel{
		Object: client.makeObject(id),
	}
}

func (*BotClient) FetchChannel(id Snowflake) Channel {
	// TODO
	return Channel{}
}

func (channel *Channel) SendMessage(content string) (*Message, error) {
	resp, err := channel.client.restRequest(METHOD_POST, fmt.Sprintf("/channels/%s/messages", channel.ID), toJSON(MessageSendData{
		Content: &content,
	}))

	if err != nil {
		return nil, err
	}

	var data Message
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}

	data.Object = channel.client.makeObject(data.ID)
	data.Available = true
	return &data, nil
}
