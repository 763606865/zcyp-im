package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"zcyp-im/internal/model"
	"zcyp-im/internal/service"
)

const writeTimeout = 5 * time.Second

type Hub struct {
	imService *service.IMService

	mu            sync.RWMutex
	clients       map[*Client]struct{}
	conversations map[string]map[*Client]struct{}
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	appCode string
	userID  string

	mu            sync.Mutex
	subscriptions map[string]struct{}
}

type inboundMessage struct {
	Action         string `json:"action"`
	ConversationNo string `json:"conversation_no"`
	MessageType    string `json:"message_type"`
	ClientMsgID    string `json:"client_msg_id"`
	Content        string `json:"content"`
}

type outboundMessage struct {
	Action         string         `json:"action"`
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

func (h *Hub) Register(conn *websocket.Conn, appCode, userID string) *Client {
	client := &Client{
		hub:           h,
		conn:          conn,
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
	h.mu.RLock()
	subscribers := h.conversations[conversationNo]
	clients := make([]*Client, 0, len(subscribers))
	for client := range subscribers {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	payload := outboundMessage{
		Action:         "message",
		ConversationNo: conversationNo,
		Message:        &message,
	}

	for _, client := range clients {
		_ = client.WriteJSON(payload)
	}
}

func (c *Client) Serve(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close(websocket.StatusNormalClosure, "bye")
	}()

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
			if msg.ConversationNo == "" || msg.MessageType == "" || msg.Content == "" {
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
			})
			if err != nil {
				_ = c.WriteJSON(outboundMessage{Action: "error", Error: err.Error()})
				continue
			}

			c.hub.BroadcastMessage(msg.ConversationNo, message)
		default:
			_ = c.WriteJSON(outboundMessage{Action: "error", Error: "unsupported action"})
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
