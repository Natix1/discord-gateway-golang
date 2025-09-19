package discord

/*

var (
	// INITIAL gateway url
	gatewayUrl = "wss://gateway.discord.gg?v=10&encoding=json"

	// Internal makes
	websocketWrites = make(chan json.RawMessage)
	exit            = make(chan int)
	eventCallbacks  = make(map[int]func(data Event))

	// Internal variables
	wsconn         *websocket.Conn = nil
	debugMode                      = true
	nextCallbackId                 = 0
	callbackMutex  sync.RWMutex

	// Reconnecting
	lastSerial    *int       = nil
	lastSessionId *Snowflake = nil
	lastResumeUrl *string    = nil

	// Bot settings
	intents        = 33281
	token   string = ""
)

*/

/*

func onDisconnect(errorCode int) {
	reconnectCodes := []int{4000, 4001, 4002, 4003, 4005, 4007, 4008, 4009}
	shouldReconnect := slices.Contains(reconnectCodes, errorCode)

	if !shouldReconnect {
		log.Fatalf("Received non-recconectable exit code from discord: %d\n", errorCode)
	}

	conn, _, err := websocket.DefaultDialer.Dial(*lastResumeUrl, nil)
	if err != nil {
		log.Fatalf("Failed to reconnect to Discord: %s", err)
	}
	wsconn = conn
	debug("Reconnected to Discord")

	reconnect()
}

*/
