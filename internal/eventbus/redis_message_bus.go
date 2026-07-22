package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"zcyp-im/internal/config"
	"zcyp-im/internal/model"
)

const publishTimeout = 3 * time.Second
const subscribeRetryDelay = time.Second

type messageEvent struct {
	ConversationNo string        `json:"conversation_no"`
	Message        model.Message `json:"message"`
}

type RedisMessageBus struct {
	client  *redis.Client
	channel string
}

func NewRedisMessageBus(cfg config.RedisConfig) *RedisMessageBus {
	return &RedisMessageBus{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		channel: cfg.Channel,
	}
}

func (b *RedisMessageBus) Ping(ctx context.Context) error {
	return b.client.Ping(ctx).Err()
}

func (b *RedisMessageBus) Close() error {
	return b.client.Close()
}

func (b *RedisMessageBus) BroadcastMessage(conversationNo string, message model.Message) {
	payload, err := json.Marshal(messageEvent{ConversationNo: conversationNo, Message: message})
	if err != nil {
		log.Printf("message bus: encode event failed conversation_no=%s err=%v", conversationNo, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()
	if err := b.client.Publish(ctx, b.channel, payload).Err(); err != nil {
		log.Printf("message bus: publish failed conversation_no=%s err=%v", conversationNo, err)
	}
}

func (b *RedisMessageBus) Subscribe(ctx context.Context, handler func(string, model.Message)) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		err := b.subscribeOnce(ctx, handler)
		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}
		log.Printf("message bus: subscription interrupted channel=%s err=%v", b.channel, err)

		timer := time.NewTimer(subscribeRetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

func (b *RedisMessageBus) subscribeOnce(ctx context.Context, handler func(string, model.Message)) error {
	pubsub := b.client.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}
	log.Printf("message bus: subscribed channel=%s", b.channel)

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			return err
		}

		var event messageEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("message bus: decode event failed err=%v", err)
			continue
		}
		if event.ConversationNo == "" || event.Message.MessageNo == "" {
			log.Printf("message bus: ignored invalid event conversation_no=%s", event.ConversationNo)
			continue
		}
		handler(event.ConversationNo, event.Message)
	}
}
