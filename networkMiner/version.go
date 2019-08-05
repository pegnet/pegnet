package networkMiner

// Include versioning. Might be helpful in the future.
const version = "0.1.0"

func NewNetworkMessage(cmd int, data interface{}) *NetworkMessage {
	return &NetworkMessage{NetworkCommand: cmd, Data: data, Version: version}
}
