# Nginx 部署

本文档说明如何通过 Nginx 反向代理 `zcyp-im` 的 API 服务和 WebSocket 网关。

当前默认端口：

- API 服务：`127.0.0.1:9011`
- WebSocket 网关：`127.0.0.1:9012`

## 1. HTTP 反向代理

适用于仅内网访问或未启用 HTTPS 的场景。

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

upstream zcyp_im_api {
    server 127.0.0.1:9011;
}

upstream zcyp_im_ws {
    server 127.0.0.1:9012;
}

server {
    listen 80;
    server_name im.example.com;

    client_max_body_size 20m;

    location /health {
        proxy_pass http://zcyp_im_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /admin/ {
        proxy_pass http://zcyp_im_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /im/connect {
        proxy_pass http://zcyp_im_ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_buffering off;
    }

    location /im/ {
        proxy_pass http://zcyp_im_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 2. HTTPS / WSS 反向代理

适用于公网正式环境。浏览器和 App 接入 WebSocket 时，建议统一使用 `wss://`。

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

upstream zcyp_im_api {
    server 127.0.0.1:9011;
}

upstream zcyp_im_ws {
    server 127.0.0.1:9012;
}

server {
    listen 80;
    server_name im.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name im.example.com;

    ssl_certificate /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    client_max_body_size 20m;

    location /health {
        proxy_pass http://zcyp_im_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /admin/ {
        proxy_pass http://zcyp_im_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /im/connect {
        proxy_pass http://zcyp_im_ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_buffering off;
    }

    location /im/ {
        proxy_pass http://zcyp_im_api;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 3. 客户端接入地址

HTTP 接口：

```text
https://im.example.com/admin/...
https://im.example.com/im/...
```

WebSocket 连接：

```text
wss://im.example.com/im/connect?token=<token>
```

## 4. 关键说明

1. `location /im/connect` 必须放在 `location /im/` 之前，避免被普通 IM API 路由覆盖。
2. WebSocket 代理必须带上以下配置，否则升级握手会失败：
   - `proxy_http_version 1.1`
   - `proxy_set_header Upgrade $http_upgrade`
   - `proxy_set_header Connection $connection_upgrade`
3. `proxy_read_timeout` 和 `proxy_send_timeout` 要足够大，否则空闲连接会被 Nginx 主动断开。
4. `proxy_buffering off` 建议保留，避免长连接事件推送被缓冲。
5. 如果你后面把 API 或 WS 端口改了，需要同步修改 `upstream` 配置。

## 5. 启动顺序建议

先启动后端服务，再重载 Nginx：

```bash
make run-api
make run-ws
sudo nginx -t
sudo nginx -s reload
```
