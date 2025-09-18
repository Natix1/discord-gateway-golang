package main

var ()

func init() {
	addEventCallback(func(data Event) {
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
