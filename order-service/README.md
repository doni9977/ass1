# order-service

Order service. Creates orders, requests payment authorization from payment-service, and stores data in PostgreSQL.

## What this service does

- POST /orders - create an order
- GET /orders/:id - get an order by ID
- PATCH /orders/:id/cancel - cancel an order (orders with status Paid cannot be canceled)

## Ports and dependencies

- order-service: http://localhost:8080
- payment-service (required to create orders): http://localhost:8081
- PostgreSQL for order-service: localhost:5434, database orderdb

Connection string hardcoded in the app:

```go
postgres://user:pass@localhost:5434/orderdb?sslmode=disable
```

## 1) Start PostgreSQL in Docker

Run from this project folder:

```bash
docker run -d \
	--name order-postgres \
	-e POSTGRES_USER=user \
	-e POSTGRES_PASSWORD=pass \
	-e POSTGRES_DB=orderdb \
	-p 5434:5432 \
	postgres:16
```

Check container status:

```bash
docker ps --filter name=order-postgres
```

## 2) Apply migration

```bash
docker exec -i order-postgres psql -U user -d orderdb < migrations/up.sql
```

Check table exists:

```bash
docker exec -it order-postgres psql -U user -d orderdb -c "\dt"
```

## 3) Start payment-service first

In a separate terminal:

```bash
cd ../payment-service
go mod tidy
go run ./cmd/payment-service
```

## 4) Start order-service

In the order-service folder:

```bash
go mod tidy
go run ./cmd/order-service
```

## Postman requests

Base URL:

```text
http://localhost:8080
```

### 1. Create order

POST /orders

Headers:

```text
Content-Type: application/json
Idempotency-Key: order-001
```

Body (raw JSON):

```json
{
	"customer_id": "cust-1",
	"item_name": "iPhone 15",
	"amount": 50000
}
```

Example response (201 Created):

```json
{
	"ID": "8d0dc7f1-1cd3-4d5c-a4ef-e9d5c575d0d3",
	"CustomerID": "cust-1",
	"ItemName": "iPhone 15",
	"Amount": 50000,
	"Status": "Paid",
	"CreatedAt": "2026-04-01T12:00:00Z",
	"IdempotencyKey": "order-001"
}
```

If payment-service is unavailable, response is 503 Service Unavailable:

```json
{
	"error": "service unavailable"
}
```

If you send the same Idempotency-Key again with the same payload, the existing order is returned (no duplicate row).

### 2. Get order by ID

GET /orders/{id}

Example:

```text
GET http://localhost:8080/orders/8d0dc7f1-1cd3-4d5c-a4ef-e9d5c575d0d3
```

Response 200 OK returns the order object.

If not found:

```json
{
	"error": "order not found"
}
```

### 3. Cancel order

PATCH /orders/{id}/cancel

Example:

```text
PATCH http://localhost:8080/orders/8d0dc7f1-1cd3-4d5c-a4ef-e9d5c575d0d3/cancel
```

Response is 200 OK with empty body.

If order status is Paid, response is 400 Bad Request:

```json
{
	"error": "cannot cancel paid order"
}
```

## Quick curl test

```bash
curl -i -X POST http://localhost:8080/orders \
	-H "Content-Type: application/json" \
	-H "Idempotency-Key: order-001" \
	-d '{"customer_id":"cust-1","item_name":"iPhone 15","amount":50000}'
```

## Stop and clean Docker

Stop and remove DB container:

```bash
docker rm -f order-postgres
```

Remove Postgres image (optional):

```bash
docker rmi postgres:16
```

## Common issues

- dial tcp ...: connect: connection refused
	- Make sure order-postgres container is running.
- relation "orders" does not exist
	- Migration migrations/up.sql was not applied.
- service unavailable when creating order
	- payment-service is not running on localhost:8081.

