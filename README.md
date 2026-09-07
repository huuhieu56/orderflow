# OrderFlow

OrderFlow là project học Go backend theo kiến trúc microservices. Bốn service được build và deploy độc lập, sở hữu PostgreSQL database riêng, giao tiếp bằng HTTP và Kafka.

| Service | Trách nhiệm |
| --- | --- |
| Auth | Register, login, access/refresh token, role và login throttling |
| Product | Product catalog, stock metadata và Redis cache-aside |
| Order | Snapshot sản phẩm, tính tổng tiền, lưu order và transactional outbox |
| Notification | Consume Kafka idempotently và quản lý notification |

Gateway là entry point public tại `${GATEWAY_PORT:-8080}`. Mỗi container backend listen cổng `8080` trong Compose network.

## Chạy hệ thống

```bash
docker compose up --build -d
docker compose ps
```

Nếu cổng 8080 đã được dùng:

```bash
GATEWAY_PORT=8090 docker compose up --build -d
```

Gateway và mỗi backend có health endpoint `/health`.

## Kiểm tra code

Repo dùng Go workspace gồm nhiều module. Chạy kiểm tra theo từng module:

```bash
for dir in libs/httpx libs/identity libs/sqlmigrate services/auth services/product services/order services/notification; do
  (cd "$dir" && go test ./... && go vet ./...)
done
```

Project chủ ý chưa có test. `go test` hiện được dùng như compile gate.

## HTTP routes

- Auth: `POST /api/v1/auth/register`, `login`, `refresh`, `logout`; `GET /api/v1/auth/me`.
- Product: `POST`, `GET /api/v1/products`; `GET`, `PATCH /api/v1/products/{id}`. Write routes yêu cầu role `admin`.
- Order: `POST`, `GET /api/v1/orders`; `GET`, `DELETE /api/v1/orders/{id}`. Tất cả yêu cầu access token.
- Notification: `GET /api/v1/notifications`; `PATCH /api/v1/notifications/{id}/read`. Tất cả yêu cầu access token.

Các field tiền dùng integer minor unit; project hiện giả định một currency duy nhất. Order chỉ kiểm tra stock tại thời điểm tạo, chưa reserve hay decrement stock.
