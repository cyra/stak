package ports

import "stak/internal/models"

// CategorizerPort defines the interface for entry categorization
type CategorizerPort interface {
	CategoriseEntry(entry *models.Entry)
}