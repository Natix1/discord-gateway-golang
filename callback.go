package discord

const (
	_OpcodeCallbackType    = 0
	_EventNameCallbackType = 1
	_AnyCallbackType       = 2
	_ReadyCallbackType     = 3
)

func (cb *Callback) Disconnect() {
	cb.client.callbackMutex.Lock()
	defer cb.client.callbackMutex.Unlock()

	delete(cb.client.callbacks, cb.id)
}
