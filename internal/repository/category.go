package repository

import (
	"database/sql"

	"forum/internal/domain"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List() ([]domain.Category, error) {
	rows, err := r.db.Query(`SELECT id, name FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) IDByName(name string) (int, error) {
	var id int
	err := r.db.QueryRow(`SELECT id FROM categories WHERE name = ?`, name).Scan(&id)
	return id, err
}
