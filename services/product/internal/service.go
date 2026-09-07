package product

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"orderflow/product/internal/cache"
)

const productsListCacheKey = "products:list:active"
const cacheTTL = 5 * time.Minute

type Service struct {
	repo  *Repository
	cache *cache.Cache
}

func NewService(repo *Repository, cache *cache.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) Create(ctx context.Context, name, description string, price, stock int64) (*Product, error) {
	name = strings.TrimSpace(name)
	if name == "" || price <= 0 || stock < 0 {
		return nil, ErrInvalidProduct
	}
	p := &Product{Name: name, Description: description, Price: price, Stock: stock, Status: StatusActive}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, productsListCacheKey)
	return p, nil
}

func (s *Service) List(ctx context.Context) ([]*Product, error) {
	if cached, found, err := s.cache.Get(ctx, productsListCacheKey); err == nil && found {
		var products []*Product
		if json.Unmarshal([]byte(cached), &products) == nil {
			return products, nil
		}
	}
	products, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(products); err == nil {
		_ = s.cache.Set(ctx, productsListCacheKey, data, cacheTTL)
	}
	return products, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Product, error) {
	key := fmt.Sprintf("products:%d", id)
	if cached, found, err := s.cache.Get(ctx, key); err == nil && found {
		var p Product
		if json.Unmarshal([]byte(cached), &p) == nil {
			return &p, nil
		}
	}
	p, err := s.repo.GetByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(p); err == nil {
		_ = s.cache.Set(ctx, key, data, cacheTTL)
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id int64, update UpdateProduct) (*Product, error) {
	if update.Name == nil &&
		update.Description == nil &&
		update.Price == nil &&
		update.Stock == nil &&
		update.Status == nil {
		return nil, ErrInvalidProduct
	}
	p, err := s.repo.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	if update.Name != nil {
		name := strings.TrimSpace(*update.Name)
		if name == "" {
			return nil, ErrInvalidProduct
		}
		p.Name = name
	}
	if update.Description != nil {
		p.Description = *update.Description
	}
	if update.Price != nil {
		if *update.Price <= 0 {
			return nil, ErrInvalidProduct
		}
		p.Price = *update.Price
	}
	if update.Stock != nil {
		if *update.Stock < 0 {
			return nil, ErrInvalidProduct
		}
		p.Stock = *update.Stock
	}
	if update.Status != nil {
		if *update.Status != StatusActive && *update.Status != StatusInactive {
			return nil, ErrInvalidProduct
		}
		p.Status = *update.Status
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, productsListCacheKey)
	_ = s.cache.Del(ctx, fmt.Sprintf("products:%d", id))
	return p, nil
}
