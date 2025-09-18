package discordgateway

func init() {
	AddEventCallback(func(data Event) {
		if data.Opcode != HelloOpcode {
			return
		}

		identifyData := IdentifyData{
			Token:   token,
			Intents: intents,
			Properties: struct {
				OperatingSystem string `json:"os"`
				Browser         string `json:"browser"`
				Device          string `json:"device"`
			}{
				OperatingSystem: "Linux",
				Browser:         "github.com/natix1/discord-gateway-golang",
				Device:          "github.com/natix1/discord-gateway-golang",
			},
		}

		eventData := Event{
			Opcode: IdentifyOpcode,
			Data:   *toJSON(identifyData),
		}

		writeToWebsocket(*toJSON(eventData))
	})
}

/*
identify(IdentifyData{
	Token:   token,
	Intents: 513,
	Properties: struct {
		OperatingSystem string "json:\"os\""
		Browser         string "json:\"browser\""
		Device          string "json:\"device\""
	}{
		OperatingSystem: "Linux",
		Browser:         "github.com/natix1/discord-gateway-golang",
		Device:          "github.com/natix1/discord-gateway-golang",
	},
})
*/
