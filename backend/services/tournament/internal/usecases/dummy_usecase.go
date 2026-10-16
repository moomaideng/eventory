package usecases

import (
	"context"
	"errors"
	"strings"

	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrTitleRequired = errors.New("title is required")
)

type DummyUseCase struct {
	repo repositories.DummyRepository
}

func NewDummyUseCase(repo repositories.DummyRepository) *DummyUseCase {
	return &DummyUseCase{repo: repo}
}

func (u *DummyUseCase) CreateDummy(ctx context.Context, title string, content string) (*models.DummyResource, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, ErrTitleRequired
	}

	doc := &models.DummyResource{
		Title:   trimmedTitle,
		Content: strings.TrimSpace(content),
	}

	return u.repo.Create(ctx, doc)
}

func (u *DummyUseCase) GetDummy(ctx context.Context, id string) (*models.DummyResource, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *DummyUseCase) ListDummies(ctx context.Context) ([]models.DummyResource, error) {
	return u.repo.List(ctx)
}
