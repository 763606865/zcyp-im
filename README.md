# zcyp-im

`zcyp-im` is an IM platform service skeleton for multi-product access.

Current scope:

- `admin` API for application provisioning
- `im` API for token issuing and message/member operations
- dedicated WebSocket gateway for long connections
- config-driven API and WebSocket bootstrap
- repository abstraction with in-memory and MySQL implementations

Run:

```bash
make run
make run-ws
```

Default routes:

- `GET /health`
- `GET /admin/apps`
- `POST /admin/apps`
- `GET /admin/apps/:app_code`
- `GET /admin/apps/:app_code/users`
- `POST /admin/apps/:app_code/users`
- `GET /admin/apps/:app_code/users/:external_user_id`
- `PATCH /admin/apps/:app_code/users/:external_user_id/status`
- `POST /im/auth/token`
- `POST /im/conversations`
- `POST /im/conversations/:conversation_no/messages`
- `GET /im/conversations/:conversation_no/messages`
- `GET /im/conversations/:conversation_no/members`
- `POST /im/conversations/:conversation_no/members`
- `DELETE /im/conversations/:conversation_no/members/:member_user_id`
- `POST /im/conversations/:conversation_no/join`
- `POST /im/conversations/:conversation_no/leave`
- `POST /im/conversations/:conversation_no/all-muted`
- `POST /im/conversations/:conversation_no/review`
- `POST /im/conversations/:conversation_no/members/:member_user_id/ban`
- `POST /im/conversations/:conversation_no/members/:member_user_id/unban`
- `POST /im/conversations/:conversation_no/members/:member_user_id/role`
- `POST /im/conversations/:conversation_no/members/:member_user_id/mic`
- `POST /im/conversations/:conversation_no/members/:member_user_id/mute`
- `POST /im/conversations/:conversation_no/members/:member_user_id/unmute`

Auth flow:

- create or update user first with `POST /admin/apps/:app_code/users`
- disable or re-enable user with `PATCH /admin/apps/:app_code/users/:external_user_id/status`
- exchange token with `POST /im/auth/token` using `app_code + app_key + user_id`
- call IM HTTP APIs with `Authorization: Bearer <token>`
- connect WebSocket with `GET ws://<ws-host>:<ws-port>/im/connect?token=<token>`

WebSocket:

- connect with `GET /im/connect?token=...` on the WebSocket gateway
- client message `{"action":"subscribe","conversation_no":"conv_xxx"}`
- client message `{"action":"send_message","conversation_no":"conv_xxx","message_type":"text","content":"hello"}`
- server event `{"action":"message","conversation_no":"conv_xxx","message":{...}}`

Conversation:

- `POST /im/conversations` body example: `{"type":"group","subject":"demo","member_user_ids":["u_2","u_3"]}`
- creator is auto-added as `owner`
- only conversation members can subscribe, send messages, and load history
- only `owner` can add or remove members
- supported types: `single`, `group`, `chatroom`, `live_room`
- `single` requires exactly one target user in `member_user_ids`
- `chatroom` and `live_room` do not accept initial `member_user_ids`
- `chatroom` and `live_room` support self join/leave
- `owner` or `admin` can ban and unban members
- `owner` or `admin` can adjust member role
- `owner` or `admin` can enable `all-muted`
- `owner` or `admin` can enable `review`
- muted members cannot send messages until unmuted or mute timeout expires
- `live_room` self-join users get `audience` role by default
- `live_room` only `owner`, `admin`, `speaker` can send messages
- role update body example: `{"role":"speaker"}`
- mic update body example: `{"mic_status":"on"}`
- mute body example: `{"minutes":30}`
- all muted body example: `{"enabled":true}`
- review body example: `{"enabled":true}`
- if `review` is enabled, messages containing configured blocked words are rejected

MySQL config:

- defaults live in `configs/config.yaml`
- environment variables override config, for example `ZCYP_IM_MYSQL_PASSWORD`
- the app auto-loads `.env` and `.env.local` on startup
- copy `.env.example` to `.env` for local setup
- keep real secrets in `.env`, do not commit them
- API default listen address: `0.0.0.0:9011`
- WebSocket default listen address: `0.0.0.0:9012`
- WebSocket config can also be overridden with `ZCYP_IM_WEBSOCKET_HOST` and `ZCYP_IM_WEBSOCKET_PORT`

Migration files:

- `migrations/000001_init.up.sql`
- `migrations/000001_init.down.sql`
- `migrations/000002_conversation_members.up.sql`
- `migrations/000002_conversation_members.down.sql`
- `migrations/000003_member_status.up.sql`
- `migrations/000003_member_status.down.sql`
- `migrations/000004_member_moderation.up.sql`
- `migrations/000004_member_moderation.down.sql`
- `migrations/000005_conversation_controls.up.sql`
- `migrations/000005_conversation_controls.down.sql`

Manual bootstrap:

```bash
mysql -uroot -p zcyp_im < migrations/000001_init.up.sql
mysql -uroot -p zcyp_im < migrations/000002_conversation_members.up.sql
mysql -uroot -p zcyp_im < migrations/000003_member_status.up.sql
mysql -uroot -p zcyp_im < migrations/000004_member_moderation.up.sql
mysql -uroot -p zcyp_im < migrations/000005_conversation_controls.up.sql
```

Recommended migration tool:

```bash
migrate -path migrations -database "mysql://root:password@tcp(127.0.0.1:3306)/zcyp_im" up
```

Common commands:

```bash
make run-api
make run-ws
make test
make build
make migrate-up
make migrate-down
make migrate-force VERSION=1
```

Docs:

- [docs/接口总览.md](./docs/接口总览.md)
- [docs/WebSocket.md](./docs/WebSocket.md)
- [docs/nginx.md](./docs/nginx.md)
