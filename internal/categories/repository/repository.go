package repository

import (
	"errors"
	"log"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/datetime"
)

type CategoryRepository interface {
	CreateCategory(category *entity.Category) error
	UpdateCategory(id int64, category *entity.Category) error
	DeleteCategory(id int64) error
	GetCategoryByID(id int64) (*entity.ResponseCategory, error)
	GetAllCategories() ([]entity.ResponseCategory, error)
}

type categoryRepository struct {
	db *database.DB
}

func NewCategoryRepository(db *database.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) CreateCategory(category *entity.Category) error {
	var (
		query string
		err   error
	)

	query = "INSERT INTO categories (name, description, created_at, updated_at) VALUES ($1, $2, $3, $4)"

	err = r.db.WithTx(func(tx *database.Tx) error {
		err = tx.WithStmt(query, func(stmt *database.Stmt) error {
			_, err = stmt.Exec(category.Name, category.Description, "now()", "now()")
			return err
		})

		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *categoryRepository) UpdateCategory(id int64, category *entity.Category) error {
	var (
		query string
		err   error
	)

	query = "UPDATE categories SET name = $1, description = $2, updated_at = $3 WHERE id = $4"

	err = r.db.WithTx(func(tx *database.Tx) error {
		err = tx.WithStmt(query, func(stmt *database.Stmt) error {
			_, err = stmt.Exec(category.Name, category.Description, "now()", id)
			return err
		})

		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *categoryRepository) DeleteCategory(id int64) error {
	var (
		query string
		err   error
	)

	query = "DELETE FROM categories WHERE id = $1"

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

func (r *categoryRepository) GetCategoryByID(id int64) (*entity.ResponseCategory, error) {
	var (
		category     entity.Category
		respCategory entity.ResponseCategory
		err          error
		query        string
	)

	query = "SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1"

	err = r.db.WithStmt(query, func(stmt *database.Stmt) error {
		err = stmt.Query(func(rows *database.Rows) error {
			if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt); err != nil {
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

	log.Println("category.CreatedAt : ", category.CreatedAt)
	log.Println("category.UpdatedAt : ", category.UpdatedAt)

	createdAt, _ := datetime.ParseTime(category.CreatedAt)
	updatedAt, _ := datetime.ParseTime(category.UpdatedAt)

	respCategory = entity.ResponseCategory{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return &respCategory, nil
}

func (r *categoryRepository) GetAllCategories() ([]entity.ResponseCategory, error) {
	var (
		categories []entity.Category
		err        error
		query      string
	)

	query = "SELECT id, name, description, created_at, updated_at FROM categories"

	err = r.db.WithStmt(query, func(stmt *database.Stmt) error {
		err = stmt.Query(func(rows *database.Rows) error {
			var category entity.Category
			if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt); err != nil {
				return err
			}

			categories = append(categories, category)
			return nil
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	var respCategories []entity.ResponseCategory
	for _, category := range categories {
		createdAt, _ := datetime.ParseTime(category.CreatedAt)
		updatedAt, _ := datetime.ParseTime(category.UpdatedAt)

		respCategory := entity.ResponseCategory{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		respCategories = append(respCategories, respCategory)
	}

	return respCategories, nil
}
