package discordgateway

var ()

func init() {
	AddEventCallback(func(data Event) {
		if data.Serial == nil {
			return
		}

		if lastSerial == nil {
			lastSerial = data.Serial
		} else {
			if *lastSerial < *data.Serial {
				lastSerial = data.Serial
			}
		}
	})
}
