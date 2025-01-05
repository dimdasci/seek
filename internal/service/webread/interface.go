package webread

import (
	"context"

	"github.com/dimdasci/seek/internal/models"
)

// WebReader defines the interface for web reading functionality.
type WebReader interface {
	Read(ctx context.Context, urls []string) (*models.WebPages, error)
}
