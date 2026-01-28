package repository

import (
	"errors"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/datetime"
)

type ProductRepository interface {
	CreateProduct(product *entity.Product) error
	UpdateProduct(id int64, product *entity.Product) error
	DeleteProduct(id int64) error
	GetProductByID(id int64) (*entity.ResponseProductWithCategories, error)
	GetAllProducts() ([]entity.ResponseProductWithCategories, error)
	GetCategoryByID(id int64) (*entity.Category, error)
}

type productRepository struct {
	db *database.DB
}

func NewProductRepository(db *database.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(product *entity.Product) error {
	var (
		query string
		err   error
	)

	query = "INSERT INTO products (name, price, stock, category_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)"

	err = r.db.WithTx(func(tx *database.Tx) error {
		err = tx.WithStmt(query, func(stmt *database.Stmt) error {
			_, err = stmt.Exec(product.Name, product.Price, product.Stock, product.CategoryID, "now()", "now()")
			return err
		})

		if err != nil {
			return err
		}

		return nil
	})

	return err

}

func (r *productRepository) UpdateProduct(id int64, product *entity.Product) error {
	var (
		query string
		err   error
	)

	query = "UPDATE products SET name = $1, price = $2, stock = $3, category_id = $4, updated_at = $5 WHERE id = $6"

	err = r.db.WithTx(func(tx *database.Tx) error {
		err = tx.WithStmt(query, func(stmt *database.Stmt) error {
			_, err := stmt.Exec(product.Name, product.Price, product.Stock, product.CategoryID, "now()", id)
			return err
		})

		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *productRepository) DeleteProduct(id int64) error {
	var (
		query string
		err   error
	)

	query = "DELETE FROM products WHERE id = $1"

	err = r.db.WithTx(func(tx *database.Tx) error {
		err = tx.WithStmt(query, func(stmt *database.Stmt) error {
			_, err = stmt.Exec(id)
			return err
		})

		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *productRepository) GetAllProducts() ([]entity.ResponseProductWithCategories, error) {
	var (
		query             string
		products          []entity.ProductWithCategories
		productCategories []entity.ResponseProductWithCategories
		err               error
	)

	query = "SELECT products.id, products.name, products.price, products.stock, products.created_at, products.updated_at, categories.id as category_id, categories.name as category_name FROM products JOIN categories ON products.category_id = categories.id"

	err = r.db.WithStmt(query, func(stmt *database.Stmt) error {
		err = stmt.Query(func(rows *database.Rows) error {
			var product entity.ProductWithCategories
			if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt, &product.CategoryID, &product.CategoryName); err != nil {
				return err
			}

			products = append(products, product)
			return nil
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	for _, product := range products {
		createdAt, _ := datetime.ParseTime(product.CreatedAt)
		updatedAt, _ := datetime.ParseTime(product.UpdatedAt)

		productCategories = append(productCategories, entity.ResponseProductWithCategories{
			ID:           product.ID,
			Name:         product.Name,
			Price:        product.Price,
			Stock:        product.Stock,
			CategoryName: product.CategoryName,
			CategoryID:   product.CategoryID,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		})
	}

	return productCategories, nil
}

func (r *productRepository) GetProductByID(id int64) (*entity.ResponseProductWithCategories, error) {
	var (
		product         entity.ProductWithCategories
		productCategory entity.ResponseProductWithCategories
		err             error
		query           string
	)

	query = "SELECT products.id, products.name, products.price, products.stock, products.created_at, products.updated_at, categories.id as category_id, categories.name as category_name FROM products JOIN categories ON products.category_id = categories.id WHERE products.id = $1"

	err = r.db.WithStmt(query, func(stmt *database.Stmt) error {
		err = stmt.Query(func(rows *database.Rows) error {
			if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt, &product.CategoryID, &product.CategoryName); err != nil {
				return err
			}

			return nil
		}, id)

		return err
	})

	if err != nil {
		return nil, err
	}

	if product.ID == 0 {
		return nil, errors.New("product not found")
	}

	createdAt, _ := datetime.ParseTime(product.CreatedAt)
	updatedAt, _ := datetime.ParseTime(product.UpdatedAt)

	productCategory = entity.ResponseProductWithCategories{
		ID:           product.ID,
		Name:         product.Name,
		Price:        product.Price,
		Stock:        product.Stock,
		CategoryID:   product.CategoryID,
		CategoryName: product.CategoryName,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	return &productCategory, nil
}

func (r *productRepository) GetCategoryByID(id int64) (*entity.Category, error) {
	var (
		category entity.Category
		err      error
		query    string
	)

	query = "SELECT id, name FROM categories WHERE id = $1"

	err = r.db.WithStmt(query, func(stmt *database.Stmt) error {
		err = stmt.Query(func(rows *database.Rows) error {
			if err := rows.Scan(&category.ID, &category.Name); err != nil {
				return err
			}

			return nil
		}, id)

		return err
	})

	if err != nil {
		return nil, err
	}

	if category.ID == 0 {
		return nil, errors.New("category not found")
	}

	return &category, nil
}
