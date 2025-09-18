package main

import (
	"encoding/json"
	"time"
)

type Snowflake string

type Event struct {
	Opcode    int             `json:"op"`
	Data      json.RawMessage `json:"d"`
	EventName *string         `json:"t"`
	Serial    *int            `json:"s"`
}

type IdentifyData struct {
	Token      string `json:"token"`
	Intents    int    `json:"intents"`
	Properties struct {
		OperatingSystem string `json:"os"`
		Browser         string `json:"browser"`
		Device          string `json:"device"`
	} `json:"properties"`
}

type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type ReadyData struct {
	ResumeGatewayURL string    `json:"resume_gateway_url"`
	SessionID        Snowflake `json:"session_id"`
}

type ResumeData struct {
	Token      string    `json:"token"`
	SessionID  Snowflake `json:"session_id"`
	LastSerial int       `json:"seq"`
}

type User struct {
	ID          Snowflake `json:"id"`
	Username    string    `json:"username"`
	Tag         string    `json:"discriminator"`
	DisplayName *string   `json:"global_name"`
	AvatarHash  *string   `json:"avatar"`
	IsBot       *bool     `json:"bot"`
}

type Embed struct {
	Title       *string    `json:"title"`
	Type        *string    `json:"type"`
	Description *string    `json:"description"`
	Url         *string    `json:"url"`
	SentAt      *time.Time `json:"timestamp"`
	Color       *int       `json:"color"`
	Footer      *string    `json:"footer"`
}

type MessageReference struct {
	Type      int        `json:"type"`
	MessageID *Snowflake `json:"message_id"`
	ChannelID *Snowflake `json:"channel_id"`
	GuildID   *Snowflake `json:"guild_id"`
}

type MessageSendData struct {
	Content          *string           `json:"content"`
	MessageReference *MessageReference `json:"message_reference"`
	Nonce            *string           `json:"nonce"`
}

type PartialChannel struct {
	ID Snowflake `json:"id"`
}

type Application struct {
	ID          Snowflake `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconHash    *string   `json:"icon"`
	Bot         *User     `json:"bot"`
}

type Message struct {
	ID        Snowflake  `json:"id"`
	GuildID   *Snowflake `json:"guild_id"`
	ChannelID Snowflake  `json:"channel_id"`
	Author    *User      `json:"author"`
	Content   string     `json:"content"`
	Nonce     *string    `json:"nonce"`
	CreatedAt time.Time  `json:"timestamp"`
	EditedAt  *time.Time `json:"edited_timestamp"`
}
