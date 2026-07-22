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

func TestHubClassifiesKnownAndUnknownConversationClients(t *testing.T) {
	hub := NewHub(nil)
	known := &Client{appID: 1, userID: "u_1", subscriptions: map[string]struct{}{"conv_1": {}}}
	unknown := &Client{appID: 1, userID: "u_2", subscriptions: map[string]struct{}{}}
	otherApp := &Client{appID: 2, userID: "u_2", subscriptions: map[string]struct{}{}}
	nonMember := &Client{appID: 1, userID: "u_3", subscriptions: map[string]struct{}{}}
	hub.clients[known] = struct{}{}
	hub.clients[unknown] = struct{}{}
	hub.clients[otherApp] = struct{}{}
	hub.clients[nonMember] = struct{}{}

	knownClients, unknownClients := hub.classifyMessageClients("conv_1", 1, []string{"u_1", "u_2"})

	if len(knownClients) != 1 || knownClients[0] != known {
		t.Fatalf("known clients = %#v", knownClients)
	}
	if len(unknownClients) != 1 || unknownClients[0] != unknown {
		t.Fatalf("unknown clients = %#v", unknownClients)
	}
}
