package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"zcyp-im/internal/model"
	"zcyp-im/internal/service"
)

const writeTimeout = 5 * time.Second
const pingInterval = 30 * time.Second
const pingTimeout = 10 * time.Second
const maxMessageBytes = 64 * 1024

type Hub struct {
	imService *service.IMService
	publisher messagePublisher

	mu            sync.RWMutex
	clients       map[*Client]struct{}
	conversations map[string]map[*Client]struct{}
}

type messagePublisher interface {
	BroadcastMessage(conversationNo string, message model.Message)
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	appID   uint64
	appCode string
	userID  string

	mu            sync.Mutex
	subscriptions map[string]struct{}
}

func (h *Hub) SetPublisher(publisher messagePublisher) {
	h.publisher = publisher
}

func (h *Hub) PublishMessage(conversationNo string, message model.Message) {
	if h.publisher != nil {
		h.publisher.BroadcastMessage(conversationNo, message)
		return
	}
	h.BroadcastMessage(conversationNo, message)
}

type inboundMessage struct {
	Action         string          `json:"action"`
	ConversationNo string          `json:"conversation_no"`
	MessageType    string          `json:"message_type"`
	ClientMsgID    string          `json:"client_msg_id"`
	Content        json.RawMessage `json:"content"`
}

type outboundMessage struct {
	Action         string         `json:"action"`
	Event          string         `json:"event,omitempty"`
	ConversationNo string         `json:"conversation_no,omitempty"`
	Message        *model.Message `json:"message,omitempty"`
	Error          string         `json:"error,omitempty"`
}

func NewHub(imService *service.IMService) *Hub {
	return &Hub{
		imService:     imService,
		clients:       make(map[*Client]struct{}),
		conversations: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Register(conn *websocket.Conn, appID uint64, appCode, userID string) *Client {
	conn.SetReadLimit(maxMessageBytes)

	client := &Client{
		hub:           h,
		conn:          conn,
		appID:         appID,
		appCode:       appCode,
		userID:        userID,
		subscriptions: make(map[string]struct{}),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	return client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client)
	for conversationNo := range client.subscriptions {
		subscribers := h.conversations[conversationNo]
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.conversations, conversationNo)
		}
	}
}

func (h *Hub) Subscribe(client *Client, conversationNo string) error {
	if _, err := h.imService.CheckMembership(client.appCode, conversationNo, client.userID); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	client.subscriptions[conversationNo] = struct{}{}
	if _, ok := h.conversations[conversationNo]; !ok {
		h.conversations[conversationNo] = make(map[*Client]struct{})
	}
	h.conversations[conversationNo][client] = struct{}{}

	return nil
}

func (h *Hub) BroadcastMessage(conversationNo string, message model.Message) {
	memberUserIDs, err := h.imService.ListActiveConversationMemberUserIDs(message.ConversationID)
	if err != nil {
		log.Printf("websocket: resolve message audience failed conversation_no=%s err=%v", conversationNo, err)
		h.broadcastToSubscribers(conversationNo, message)
		return
	}

	knownClients, unknownClients := h.classifyMessageClients(conversationNo, message.AppID, memberUserIDs)

	payload := outboundMessage{
		Action:         "message",
		ConversationNo: conversationNo,
		Message:        &message,
	}

	for _, client := range knownClients {
		_ = client.WriteJSON(payload)
	}

	changedPayload := outboundMessage{
		Action:         "conversation_changed",
		Event:          "new_message",
		ConversationNo: conversationNo,
	}
	for _, client := range unknownClients {
		_ = client.WriteJSON(changedPayload)
	}
}

func (h *Hub) classifyMessageClients(conversationNo string, appID uint64, memberUserIDs []string) ([]*Client, []*Client) {
	members := make(map[string]struct{}, len(memberUserIDs))
	for _, userID := range memberUserIDs {
		members[userID] = struct{}{}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	knownClients := make([]*Client, 0)
	unknownClients := make([]*Client, 0)
	for client := range h.clients {
		if client.appID != appID {
			continue
		}
		if _, ok := members[client.userID]; !ok {
			continue
		}
		if _, subscribed := client.subscriptions[conversationNo]; subscribed {
			knownClients = append(knownClients, client)
		} else {
			unknownClients = append(unknownClients, client)
		}
	}
	return knownClients, unknownClients
}

func (h *Hub) broadcastToSubscribers(conversationNo string, message model.Message) {
	h.mu.RLock()
	subscribers := h.conversations[conversationNo]
	clients := make([]*Client, 0, len(subscribers))
	for client := range subscribers {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	payload := outboundMessage{Action: "message", ConversationNo: conversationNo, Message: &message}
	for _, client := range clients {
		_ = client.WriteJSON(payload)
	}
}

func (c *Client) Serve(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		c.hub.Unregister(c)
		_ = c.conn.Close(websocket.StatusNormalClosure, "bye")
	}()

	go c.pingLoop(ctx, cancel)

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}

		var msg inboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			_ = c.WriteJSON(outboundMessage{Action: "error", Error: "invalid json payload"})
			continue
		}

		switch msg.Action {
		case "subscribe":
			if msg.ConversationNo == "" {
				_ = c.WriteJSON(outboundMessage{Action: "error", Error: "conversation_no is required"})
				continue
			}
			if err := c.hub.Subscribe(c, msg.ConversationNo); err != nil {
				_ = c.WriteJSON(outboundMessage{Action: "error", Error: err.Error()})
				continue
			}
			_ = c.WriteJSON(outboundMessage{Action: "subscribed", ConversationNo: msg.ConversationNo})
		case "send_message":
			if msg.ConversationNo == "" || msg.MessageType == "" || len(msg.Content) == 0 {
				_ = c.WriteJSON(outboundMessage{Action: "error", Error: "conversation_no, message_type and content are required"})
				continue
			}

			if err := c.hub.Subscribe(c, msg.ConversationNo); err != nil {
				_ = c.WriteJSON(outboundMessage{Action: "error", Error: err.Error()})
				continue
			}

			message, err := c.hub.imService.SendMessage(service.SendMessageInput{
				AppCode:        c.appCode,
				ConversationNo: msg.ConversationNo,
				SenderUserID:   c.userID,
				MessageType:    msg.MessageType,
				ClientMsgID:    msg.ClientMsgID,
				Content:        msg.Content,
				Source:         service.SendSourceWebSocket,
			})
			if err != nil {
				_ = c.WriteJSON(outboundMessage{Action: "error", Error: err.Error()})
				continue
			}

			c.hub.PublishMessage(msg.ConversationNo, message)
		default:
			_ = c.WriteJSON(outboundMessage{Action: "error", Error: "unsupported action"})
		}
	}
}

func (c *Client) pingLoop(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, pingTimeout)
			err := c.conn.Ping(pingCtx)
			pingCancel()
			if err != nil {
				cancel()
				return
			}
		}
	}
}

func (c *Client) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	return c.conn.Write(ctx, websocket.MessageText, mustMarshal(v))
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"action":"error","error":"marshal failure"}`)
	}
	return data
}
