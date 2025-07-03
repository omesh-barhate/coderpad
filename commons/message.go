package commons

import (
	"github.com/google/uuid"
	"github.com/omesh-barhate/coderpad/crdt"
)

type Message struct {
	Username    string        `json:"username"`
	Text        string        `json:"text"`
	MessageType MessageType   `json:"type"`
	ClientID    uuid.UUID     `json:"ID"`
	Operation   Operation     `json:"operation"`
	Document    crdt.Document `json:"document"`
}

type MessageType string

const (
	DocSyncMessage MessageType = "docSync"
	DocReqMessage  MessageType = "docReq"
	SiteIDMessage  MessageType = "SiteID"
	JoinMessage    MessageType = "join"
)
