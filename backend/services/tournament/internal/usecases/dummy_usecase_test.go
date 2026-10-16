package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockDummyRepository struct {
	createFn  func(ctx context.Context, doc *models.DummyResource) (*models.DummyResource, error)
	getByIDFn func(ctx context.Context, id string) (*models.DummyResource, error)
	listFn    func(ctx context.Context) ([]models.DummyResource, error)
}

func (m *mockDummyRepository) Create(ctx context.Context, doc *models.DummyResource) (*models.DummyResource, error) {
	if m.createFn != nil {
		return m.createFn(ctx, doc)
	}
	return doc, nil
}

func (m *mockDummyRepository) GetByID(ctx context.Context, id string) (*models.DummyResource, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, repositories.ErrDummyNotFound
}

func (m *mockDummyRepository) List(ctx context.Context) ([]models.DummyResource, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return []models.DummyResource{}, nil
}

func TestDummyUseCase_CreateDummy(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedOID := bson.NewObjectID()
		repo := &mockDummyRepository{
			createFn: func(ctx context.Context, doc *models.DummyResource) (*models.DummyResource, error) {
				doc.ID = expectedOID
				doc.CreatedAt = time.Now()
				doc.UpdatedAt = time.Now()
				return doc, nil
			},
		}
		uc := usecases.NewDummyUseCase(repo)

		res, err := uc.CreateDummy(ctx, "Test Title", "Test Content")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ID != expectedOID {
			t.Errorf("expected ID %v, got %v", expectedOID, res.ID)
		}
		if res.Title != "Test Title" {
			t.Errorf("expected title 'Test Title', got %s", res.Title)
		}
	})

	t.Run("empty title returns error", func(t *testing.T) {
		repo := &mockDummyRepository{}
		uc := usecases.NewDummyUseCase(repo)

		_, err := uc.CreateDummy(ctx, "   ", "Content")
		if !errors.Is(err, usecases.ErrTitleRequired) {
			t.Fatalf("expected ErrTitleRequired, got %v", err)
		}
	})
}

func TestDummyUseCase_GetDummy(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		sampleOID := bson.NewObjectID()
		repo := &mockDummyRepository{
			getByIDFn: func(ctx context.Context, id string) (*models.DummyResource, error) {
				return &models.DummyResource{
					ID:      sampleOID,
					Title:   "Sample",
					Content: "Sample Content",
				}, nil
			},
		}
		uc := usecases.NewDummyUseCase(repo)

		res, err := uc.GetDummy(ctx, sampleOID.Hex())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ID != sampleOID {
			t.Errorf("expected ID %v, got %v", sampleOID, res.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockDummyRepository{
			getByIDFn: func(ctx context.Context, id string) (*models.DummyResource, error) {
				return nil, repositories.ErrDummyNotFound
			},
		}
		uc := usecases.NewDummyUseCase(repo)

		_, err := uc.GetDummy(ctx, "nonexistent")
		if !errors.Is(err, repositories.ErrDummyNotFound) {
			t.Fatalf("expected ErrDummyNotFound, got %v", err)
		}
	})
}
