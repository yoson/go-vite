package send_explorer

import "encoding/json"

type message struct {
	MsgType string `json:"type"`
	Data    string `json:"data"`
	Version string `json:"version"`
}

func (m *message) ToJson() ([]byte, error) {
	result, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return result, nil
}
