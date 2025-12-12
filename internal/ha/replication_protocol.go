package ha

import (
	"encoding/json"
)

type ReplicationMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
	SourceID  string                 `json:"source_id"`
}

type ReplicationProtocol struct{}

func NewReplicationProtocol() *ReplicationProtocol {
	return &ReplicationProtocol{}
}

func (rp *ReplicationProtocol) CreateDiff(state interface{}) (*ReplicationMessage, error) {
	msg := &ReplicationMessage{
		Type: "state_diff",
		Data: make(map[string]interface{}),
	}

	return msg, nil
}

func (rp *ReplicationProtocol) ApplyDiff(msg *ReplicationMessage) error {
	return nil
}

func (rp *ReplicationProtocol) Serialize(msg *ReplicationMessage) ([]byte, error) {
	return json.Marshal(msg)
}

func (rp *ReplicationProtocol) Deserialize(data []byte) (*ReplicationMessage, error) {
	var msg ReplicationMessage
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
