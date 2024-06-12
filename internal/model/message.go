package model

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	Queue string
	TxId  uint64
	Body  []byte
}

func NewMessage(queue string, userId uint64, m proto.Message) *Message {
	body, _ := proto.Marshal(m)
	return &Message{
		Queue: queue,
		TxId:  uint64(time.Now().UTC().UnixMilli())*1e6 + userId,
		Body:  body,
	}
}
