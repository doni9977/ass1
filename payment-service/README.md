# payment-service

Payment service. Accepts payment requests for orders, stores payment records, and returns authorization status.

## What this service does

- POST /payments - process a payment
- GET /payments/:order_id - get payment status by order_id

Status logic:

- amount <= 100000 -> Authorized
- amount > 100000 -> Declined

## Ports and dependencies

- payment-service: http://localhost:8081
- PostgreSQL for payment-service: localhost:5435, database paymentdb

Connection string hardcoded in the app:

```go
postgres://user:pass@localhost:5435/paymentdb?sslmode=disable
```

## 1) Start PostgreSQL in Docker

Run from this project folder:

```bash
docker run -d \
	--name payment-postgres \
	-e POSTGRES_USER=user \
	-e POSTGRES_PASSWORD=pass \
	-e POSTGRES_DB=paymentdb \
	-p 5435:5432 \
	postgres:16
```

Check container status:

```bash
docker ps --filter name=payment-postgres
```

## 2) Apply migration

```bash
docker exec -i payment-postgres psql -U user -d paymentdb < migrations/up.sql
```

Check table exists:

```bash
docker exec -it payment-postgres psql -U user -d paymentdb -c "\dt"
```

## 3) Start the service

In the payment-service folder:

```bash
go mod tidy
go run ./cmd/payment-service
```

## Postman requests

Base URL:

```text
http://localhost:8081
```

### 1. Create payment

POST /payments

Headers:

```text
Content-Type: application/json
```

Body (raw JSON):

```json
{
	"order_id": "order-100",
	"amount": 50000
}
```

Example response (200 OK):

```json
{
	"ID": "cf0df34f-a2cd-4f9a-b7e7-8c4e6fda6a63",
	"OrderID": "order-100",
	"TransactionID": "b3f9c5bb-0918-48f4-8ef3-11fef3f5bd9d",
	"Amount": 50000,
	"Status": "Authorized"
}
```

If amount is greater than 100000, status will be Declined.

If amount <= 0, response is 400 Bad Request:

```json
{
	"error": "invalid amount"
}
```

### 2. Get payment status by order_id

GET /payments/{order_id}

Example:

```text
GET http://localhost:8081/payments/order-100
```

Response 200 OK returns the payment object.

If not found:

```json
{
	"error": "payment not found"
}
```

## Quick curl test

```bash
curl -i -X POST http://localhost:8081/payments \
	-H "Content-Type: application/json" \
	-d '{"order_id":"order-100","amount":50000}'
```

## Stop and clean Docker

Stop and remove DB container:

```bash
docker rm -f payment-postgres
```

Remove Postgres image (optional):

```bash
docker rmi postgres:16
```

## Common issues

- dial tcp ...: connect: connection refused
	- Make sure payment-postgres container is running.
- relation "payments" does not exist
	- Migration migrations/up.sql was not applied.
