# WebSocket

## 1. 连接地址

```text
GET ws://<ws-host>:<ws-port>/im/connect?token=<token>
```

### 说明

- 必须先通过 `POST /im/auth/token` 获取 token
- 默认 WebSocket 网关监听 `9012` 端口，可通过 `configs/config.yaml` 中的 `websocket.port` 调整
- 连接前会再次校验用户是否存在且状态为 `active`
- 该接口必须由 WebSocket 客户端发起升级握手，不能直接用普通浏览器地址栏或普通 HTTP GET 测试
- 如果不是 WebSocket 升级请求，接口会返回：`websocket upgrade required`

## 2. 客户端消息格式

### 2.1 订阅会话

```json
{
  "action": "subscribe",
  "conversation_no": "conv_xxxxxxxx"
}
```

### 2.2 发送消息

```json
{
  "action": "send_message",
  "conversation_no": "conv_xxxxxxxx",
  "message_type": "text",
  "client_msg_id": "client_001",
  "content": {
    "text": "hello"
  }
}
```

发送业务卡片消息示例：

```json
{
  "action": "send_message",
  "conversation_no": "conv_xxxxxxxx",
  "message_type": "biz_card",
  "client_msg_id": "client_002",
  "content": {
    "card_type": "interview_invitation",
    "title": "面试邀请",
    "summary": "你收到一条新的面试邀请",
    "status": "scheduled",
    "action_url": "/recruit/interviews/67890",
    "biz_id": "interview_67890"
  }
}
```

## 3. 服务端消息格式

### 3.1 订阅成功

```json
{
  "action": "subscribed",
  "conversation_no": "conv_xxxxxxxx"
}
```

### 3.2 广播消息

```json
{
  "action": "message",
  "conversation_no": "conv_xxxxxxxx",
  "message": {
    "id": 1,
    "message_no": "msg_xxxxxxxx",
    "app_id": 1,
    "conversation_id": 1,
    "sender_user_id": "u_1001",
    "message_type": "text",
    "client_msg_id": "client_001",
    "content": {
      "type": "text",
      "text": "hello"
    },
    "created_at": "2026-07-08T15:30:00+08:00"
  }
}
```

### 3.3 错误消息

```json
{
  "action": "error",
  "error": "conversation access denied"
}
```

## 4. 连接限制

- 只有会话成员能订阅
- 发送消息会复用 HTTP 消息发送链路
- 会校验：
  - 用户状态
  - 成员状态
  - 角色权限
  - 禁言
  - 全员禁言
  - 审核开关

## 5. 跨进程实时广播

HTTP API 和 WebSocket 网关为独立进程时，必须启用 Redis 消息总线：

```env
ZCYP_IM_REDIS_ENABLED=true
ZCYP_IM_REDIS_ADDRESS=127.0.0.1:6379
ZCYP_IM_REDIS_PASSWORD=
ZCYP_IM_REDIS_DB=0
ZCYP_IM_REDIS_CHANNEL=zcyp-im:messages
```

HTTP 服务在消息落库后发布事件，所有 WebSocket 网关实例订阅事件并广播给本机已订阅该会话的连接。API 与 WebSocket 实例必须使用相同的 Redis 地址、DB 和频道。

未启用 Redis 时使用进程内广播，只适用于 HTTP 路由和 WebSocket 连接共享同一个 Hub 的单进程模式；当前独立运行 `make run-api` 与 `make run-ws` 的部署必须启用 Redis。
