package product

import (
      "errors"
      "orderflow/internal/models"
)

type Service struct {
      repo *Repository
}

func NewService(repo *Repository) *Service {
      return &Service{repo: repo}
}

func (s *Service) Create(name, description string, price float64, stock int64) (*models.Product, error) {
      if name == "" {
              return nil, errors.New("name is required")
      }
      if price < 0 {
              return nil, errors.New("price cannot be negative")
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
      return p, nil
}

func (s *Service) List(onlyActive bool) ([]*models.Product, error) {
      return s.repo.List(onlyActive)
}