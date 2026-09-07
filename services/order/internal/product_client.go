package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ProductDetails struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Price  int64  `json:"price"`
	Stock  int64  `json:"stock"`
	Status string `json:"status"`
}
type ProductFinder interface {
	GetProduct(context.Context, int64) (*ProductDetails, error)
}
type ProductClient struct {
	baseURL string
	client  *http.Client
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 3 * time.Second}}
}
func (c *ProductClient) GetProduct(ctx context.Context, id int64) (*ProductDetails, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/products/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request product service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrProductUnavailable
	}
	var envelope struct {
		Success bool           `json:"success"`
		Data    ProductDetails `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode product response: %w", err)
	}
	if !envelope.Success {
		return nil, ErrProductUnavailable
	}
	return &envelope.Data, nil
}
