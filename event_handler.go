package discordgateway

func (eh *EventDispatcher) Disconnect() {
	eh.client.callbackMutex.Lock()
	defer eh.client.callbackMutex.Unlock()

	switch eh.eventType {
	case 0:
		opcode, ok := eh.eventNameOrOpcode.(int)
		if !ok {
			return
		}

		callbacks := eh.client.opcodeCallbacks[opcode]
		if eh.id >= 0 && eh.id < len(callbacks) {
			eh.client.opcodeCallbacks[opcode] = append(callbacks[:eh.id], callbacks[eh.id+1:]...)
		}

	case 1:
		name, ok := eh.eventNameOrOpcode.(string)
		if !ok {
			return
		}

		callbacks := eh.client.eventCallbacks[name]
		if eh.id >= 0 && eh.id < len(callbacks) {
			eh.client.eventCallbacks[name] = append(callbacks[:eh.id], callbacks[eh.id+1:]...)
		}
	case 2:
		callbacks := eh.client.anyEventCalblacks
		if eh.id >= 0 && eh.id < len(callbacks) {
			eh.client.anyEventCalblacks = append(callbacks[:eh.id], callbacks[eh.id+1:]...)
		}
	}
}
