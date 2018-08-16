package send_explorer

import (
	"encoding/json"
	"math/big"
)

type message struct {
	MsgType string   `json:"type"`
	Data    string   `json:"data"`
	Version string   `json:"version"`
	EventId *big.Int `json:"eventId"`
}

func (m *message) ToJson() ([]byte, error) {
	result, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return result, nil
}
