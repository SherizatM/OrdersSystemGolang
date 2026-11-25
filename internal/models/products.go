package models

type Products struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Price       int    `json:"price" db:"price"`
	Stock       int    `json:"stock" db:"stock"`
}
