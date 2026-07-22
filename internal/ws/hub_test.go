package ws

import (
	"testing"

	"zcyp-im/internal/model"
)

type recordingPublisher struct {
	conversationNo string
	message        model.Message
}

func (p *recordingPublisher) BroadcastMessage(conversationNo string, message model.Message) {
	p.conversationNo = conversationNo
	p.message = message
}

func TestHubPublishMessageUsesConfiguredPublisher(t *testing.T) {
	hub := NewHub(nil)
	publisher := &recordingPublisher{}
	hub.SetPublisher(publisher)

	message := model.Message{MessageNo: "msg_1"}
	hub.PublishMessage("conv_1", message)

	if publisher.conversationNo != "conv_1" {
		t.Fatalf("conversation_no = %q", publisher.conversationNo)
	}
	if publisher.message.MessageNo != "msg_1" {
		t.Fatalf("message_no = %q", publisher.message.MessageNo)
	}
}
