package service

import (
	"errors"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/repository"
)

type categoryService struct {
	categoryRepository repository.CategoryRepository
}

type CategoryService interface {
	CreateCategory(requestCategory *entity.RequestCategory) error
	UpdateCategory(id int64, requestCategory *entity.RequestCategory) error
	DeleteCategory(id int64) error
	GetCategoryByID(id int64) (*entity.ResponseCategory, error)
	GetAllCategories() ([]entity.ResponseCategory, error)
	API() entity.HealthCheck
}

func NewCategoryService(categoryRepository repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepository: categoryRepository}
}

func (s *categoryService) API() entity.HealthCheck {
	return entity.HealthCheck{
		Name:      "Categories API",
		IsHealthy: true,
	}
}

func (s *categoryService) CreateCategory(requestCategory *entity.RequestCategory) error {
	category := &entity.Category{
		Name:        requestCategory.Name,
		Description: requestCategory.Description,
	}
	return s.categoryRepository.CreateCategory(category)
}

func (s *categoryService) UpdateCategory(id int64, requestCategory *entity.RequestCategory) error {
	_, err := s.categoryRepository.GetCategoryByID(id)
	if err != nil {
		return errors.New("category not found")
	}

	category := &entity.Category{
		Name:        requestCategory.Name,
		Description: requestCategory.Description,
	}
	return s.categoryRepository.UpdateCategory(id, category)
}

func (s *categoryService) DeleteCategory(id int64) error {
	_, err := s.categoryRepository.GetCategoryByID(id)
	if err != nil {
		return errors.New("category not found")
	}

	return s.categoryRepository.DeleteCategory(id)
}

func (s *categoryService) GetCategoryByID(id int64) (*entity.ResponseCategory, error) {
	return s.categoryRepository.GetCategoryByID(id)
}

func (s *categoryService) GetAllCategories() ([]entity.ResponseCategory, error) {
	return s.categoryRepository.GetAllCategories()
}
