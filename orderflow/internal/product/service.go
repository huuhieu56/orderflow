package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"orderflow/internal/cache"
	"orderflow/internal/models"
	"time"
)

const productsListCacheKey = "products:list:active"
const cacheTTL = 5 * time.Minute

type Service struct {
	repo  *Repository
	cache *cache.Cache
}

func NewService(repo *Repository, cacheSvc *cache.Cache) *Service {
	return &Service{repo: repo, cache: cacheSvc}
}

func (s *Service) Create(name, description string, price float64, stock int64) (*models.Product, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}
	if price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}

	p := &models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Status:      "active",
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	s.cache.Del(context.Background(), productsListCacheKey)

	return p, nil
}

func (s *Service) List() ([]*models.Product, error) {
	ctx := context.Background()

	// Find in Redis first
	if cached, found, err := s.cache.Get(ctx, productsListCacheKey); err == nil && found {
		var products []*models.Product
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil // Cache hit - no reach DB
		}
	}

	// Cache miss -> read DB
	products, err := s.repo.List(true)
	if err != nil {
		return nil, err
	}

	// Save to Redis
	if data, err := json.Marshal(products); err == nil {
		s.cache.Set(ctx, productsListCacheKey, data, cacheTTL)
	}

	return products, nil
}

func (s *Service) GetByID(id int64) (*models.Product, error) {
	key := fmt.Sprintf("products:%d", id)

	// Cache-Aside
	if cached, found, err := s.cache.Get(context.Background(), key); err == nil && found {
		var p models.Product
		if err := json.Unmarshal([]byte(cached), &p); err == nil {
			return &p, nil
		}
	}

	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(p); err == nil {
		s.cache.Set(context.Background(), key, data, cacheTTL)
	}

	return p, nil
}
