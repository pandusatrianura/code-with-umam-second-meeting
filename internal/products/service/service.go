package service

import (
	"errors"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/repository"
)

type productService struct {
	productRepository repository.ProductRepository
}

type ProductService interface {
	CreateProduct(product *entity.RequestProduct) error
	UpdateProduct(id int64, product *entity.RequestProduct) error
	DeleteProduct(id int64) error
	GetProductByID(id int64) (*entity.ResponseProductWithCategories, error)
	GetAllProducts() ([]entity.ResponseProductWithCategories, error)
	API() entity.HealthCheck
}

func NewProductService(productRepository repository.ProductRepository) *productService {
	return &productService{productRepository: productRepository}
}

func (s *productService) API() entity.HealthCheck {
	return entity.HealthCheck{
		Name:      "Products API",
		IsHealthy: true,
	}
}

func (s *productService) CreateProduct(requestProduct *entity.RequestProduct) error {
	_, err := s.productRepository.GetCategoryByID(int64(requestProduct.CategoryID))
	if err != nil {
		return errors.New("category not found")
	}

	product := &entity.Product{
		Name:       requestProduct.Name,
		Price:      requestProduct.Price,
		Stock:      requestProduct.Stock,
		CategoryID: requestProduct.CategoryID,
	}

	return s.productRepository.CreateProduct(product)
}

func (s *productService) UpdateProduct(id int64, requestProduct *entity.RequestProduct) error {
	_, err := s.productRepository.GetProductByID(id)
	if err != nil {
		return errors.New("product not found")
	}

	_, err = s.productRepository.GetCategoryByID(int64(requestProduct.CategoryID))
	if err != nil {
		return errors.New("category not found")
	}

	product := &entity.Product{
		Name:       requestProduct.Name,
		Price:      requestProduct.Price,
		Stock:      requestProduct.Stock,
		CategoryID: requestProduct.CategoryID,
	}

	return s.productRepository.UpdateProduct(id, product)
}

func (s *productService) DeleteProduct(id int64) error {
	_, err := s.productRepository.GetProductByID(id)
	if err != nil {
		return errors.New("product not found")
	}

	return s.productRepository.DeleteProduct(id)
}

func (s *productService) GetProductByID(id int64) (*entity.ResponseProductWithCategories, error) {
	result, err := s.productRepository.GetProductByID(id)
	return result, err
}

func (s *productService) GetAllProducts() ([]entity.ResponseProductWithCategories, error) {
	return s.productRepository.GetAllProducts()
}
