package commons

type Operation struct {
	OperationType string `json:"type"`

	Position int `json:"position"`

	Value string `json:"value"`
}
