// order-service/internal/usecase/usecase.go

package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type OrderUseCase struct {
	repo           domain.OrderRepository
	paymentGateway domain.PaymentGateway
	subscribers    map[string][]chan string
	mu             sync.RWMutex
	redisClient    *redis.Client
}

func NewOrderUseCase(repo *repository.PostgresOrderRepository, pg domain.PaymentGateway, rdb *redis.Client) *OrderUseCase {
	return &OrderUseCase{
		repo:           repo,
		paymentGateway: pg,
		subscribers:    make(map[string][]chan string),
		redisClient:    rdb,
	}
}

func (u *OrderUseCase) notifySubscribers(orderID, status string) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if chans, ok := u.subscribers[orderID]; ok {
		for _, ch := range chans {
			ch <- status
		}
	}
}

func (u *OrderUseCase) Subscribe(orderID string) chan string {
	ch := make(chan string, 10)
	u.mu.Lock()
	u.subscribers[orderID] = append(u.subscribers[orderID], ch)
	u.mu.Unlock()
	return ch
}

func (u *OrderUseCase) Unsubscribe(orderID string, ch chan string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if chans, ok := u.subscribers[orderID]; ok {
		for i, c := range chans {
			if c == ch {
				u.subscribers[orderID] = append(chans[:i], chans[i+1:]...)
				close(ch)
				break
			}
		}
	}
}

func (u *OrderUseCase) updateOrderStatusAndNotify(orderID, status string) {
	u.repo.UpdateStatus(orderID, status)
	u.notifySubscribers(orderID, status)

	ctx := context.Background()
	u.redisClient.Del(ctx, "order:"+orderID)
}

func (u *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	if idempotencyKey != "" {
		existingOrder, err := u.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil && existingOrder != nil {
			return existingOrder, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}

	if err := u.repo.Create(order); err != nil {
		return nil, err
	}
	u.notifySubscribers(order.ID, "Pending")

	status, err := u.paymentGateway.AuthorizePayment(order.ID, order.Amount)
	if err != nil {
		u.updateOrderStatusAndNotify(order.ID, "Failed")
		return nil, err
	}

	if status == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}

	u.updateOrderStatusAndNotify(order.ID, order.Status)

	return order, nil
}

func (u *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	ctx := context.Background()
	cacheKey := "order:" + id

	val, err := u.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var order domain.Order
		if err := json.Unmarshal([]byte(val), &order); err == nil {
			return &order, nil
		}
	}

	order, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	orderBytes, _ := json.Marshal(order)
	u.redisClient.Set(ctx, cacheKey, orderBytes, 5*time.Minute)

	return order, nil
}

func (u *OrderUseCase) CancelOrder(id string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status == "Paid" {
		return errors.New("cannot cancel paid order")
	}
	u.updateOrderStatusAndNotify(id, "Cancelled")
	return nil
}

func (u *OrderUseCase) GetOrdersByAmountRange(minAmountStr, maxAmountStr string) ([]*domain.Order, error) {
	if minAmountStr == "" && maxAmountStr == "" {
		return nil, errors.New("missing min_amount and max_amount")
	}

	if minAmountStr == "" || maxAmountStr == "" {
		return nil, errors.New("both min_amount and max_amount are required")
	}

	minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid min_amount")
	}

	maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid max_amount")
	}

	if minAmount < 0 {
		return nil, errors.New("min_amount less than 0")
	}

	if maxAmount > 1000000 {
		return nil, errors.New("max_amount greater than 1000000")
	}

	return u.repo.GetOrdersByAmountRange(minAmount, maxAmount)
}

func (u *OrderUseCase) ListPayments(ctx context.Context, status string) (interface{}, error) {
	return u.paymentGateway.ListPayments(ctx, status)
}

func (u *OrderUseCase) GetOrdersByAmountRange(minAmountStr, maxAmountStr string) ([]*domain.Order, error) {
	if minAmountStr == "" && maxAmountStr == "" {
		return nil, errors.New("missing min_amount and max_amount")
	}

	if minAmountStr == "" || maxAmountStr == "" {
		return nil, errors.New("both min_amount and max_amount are required")
	}

	minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid min_amount")
	}

	maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid max_amount")
	}

	if minAmount < 0 {
		return nil, errors.New("min_amount less than 0")
	}

	if maxAmount > 1000000 {
		return nil, errors.New("max_amount greater than 1000000")
	}

	return u.repo.GetOrdersByAmountRange(minAmount, maxAmount)
}
