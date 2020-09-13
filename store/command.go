package store

var (
	Set    = "set"
	Delete = "delete"
)

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func NewSetCommand(key, value string) *command {
	return &command{
		Op:    Set,
		Key:   key,
		Value: value,
	}
}

func NewDeleteCommand(key string) *command {
	return &command{
		Op:  Delete,
		Key: key,
	}
}
