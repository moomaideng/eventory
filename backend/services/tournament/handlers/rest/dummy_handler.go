package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DummyResponse struct {
	ID        bson.ObjectID `json:"id" doc:"MongoDB ObjectID hex string"`
	Title     string             `json:"title" doc:"Dummy title"`
	Content   string             `json:"content" doc:"Dummy content"`
	CreatedAt time.Time          `json:"createdAt" doc:"Created timestamp"`
	UpdatedAt time.Time          `json:"updatedAt" doc:"Updated timestamp"`
}

func toDummyResponse(doc *models.DummyResource) DummyResponse {
	return DummyResponse{
		ID:        doc.ID,
		Title:     doc.Title,
		Content:   doc.Content,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

type CreateDummyRequest struct {
	Title   string `json:"title" minLength:"1" maxLength:"120" doc:"Title of the dummy resource"`
	Content string `json:"content" maxLength:"1000" doc:"Optional content or notes"`
}

type CreateDummyInput struct {
	Body CreateDummyRequest
}

type CreateDummyOutput struct {
	Status int `default:"201"`
	Body   DummyResponse
}

type GetDummyInput struct {
	ID string `path:"id" doc:"Dummy resource ID"`
}

type GetDummyOutput struct {
	Body DummyResponse
}

type ListDummiesOutput struct {
	Body []DummyResponse
}

func RegisterDummyRoutes(api huma.API, dummyUseCase *usecases.DummyUseCase) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-dummy-resource",
		Method:        http.MethodPost,
		Path:          "/api/v1/tournaments/dummy",
		Summary:       "Create Dummy Resource (MongoDB Example)",
		Description:   "Stores a dummy document in the secondary MongoDB database",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *CreateDummyInput) (*CreateDummyOutput, error) {
		doc, err := dummyUseCase.CreateDummy(ctx, input.Body.Title, input.Body.Content)
		if err != nil {
			if errors.Is(err, usecases.ErrTitleRequired) {
				return nil, huma.Error400BadRequest("Title is required")
			}
			return nil, huma.Error500InternalServerError("Failed to create dummy resource", err)
		}
		return &CreateDummyOutput{
			Status: http.StatusCreated,
			Body:   toDummyResponse(doc),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-dummy-resource",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/dummy/{id}",
		Summary:     "Get Dummy Resource (MongoDB Example)",
		Description: "Retrieves a dummy document from MongoDB by its ID",
	}, func(ctx context.Context, input *GetDummyInput) (*GetDummyOutput, error) {
		doc, err := dummyUseCase.GetDummy(ctx, input.ID)
		if err != nil {
			if errors.Is(err, repositories.ErrDummyNotFound) || errors.Is(err, repositories.ErrInvalidID) {
				return nil, huma.Error404NotFound("Dummy resource not found")
			}
			return nil, huma.Error500InternalServerError("Failed to retrieve dummy resource", err)
		}
		return &GetDummyOutput{
			Body: toDummyResponse(doc),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-dummy-resources",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/dummy",
		Summary:     "List Dummy Resources (MongoDB Example)",
		Description: "Lists all dummy documents stored in MongoDB",
	}, func(ctx context.Context, input *struct{}) (*ListDummiesOutput, error) {
		docs, err := dummyUseCase.ListDummies(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list dummy resources", err)
		}
		items := make([]DummyResponse, len(docs))
		for i := range docs {
			items[i] = toDummyResponse(&docs[i])
		}
		return &ListDummiesOutput{
			Body: items,
		}, nil
	})
}
