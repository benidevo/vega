package home

import (
	"context"
)

// homeService defines the interface for home page data operations
type homeService interface {
	GetHomePageData(ctx context.Context, userID int, username string) (*HomePageData, error)
}
