package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/moomaideng/eventory/services/tournament/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrDummyNotFound = errors.New("dummy resource not found")
	ErrInvalidID     = errors.New("invalid dummy resource id")
)

const dummyCollectionName = "dummy_resources"

type DummyRepository interface {
	Create(ctx context.Context, doc *models.DummyResource) (*models.DummyResource, error)
	GetByID(ctx context.Context, id string) (*models.DummyResource, error)
	List(ctx context.Context) ([]models.DummyResource, error)
}

type dummyRepositoryImpl struct {
	collection *mongo.Collection
}

func NewDummyRepository(db *mongo.Database) DummyRepository {
	return &dummyRepositoryImpl{
		collection: db.Collection(dummyCollectionName),
	}
}

func (r *dummyRepositoryImpl) Create(ctx context.Context, doc *models.DummyResource) (*models.DummyResource, error) {
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	res, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		doc.ID = oid
	}

	return doc, nil
}

func (r *dummyRepositoryImpl) GetByID(ctx context.Context, id string) (*models.DummyResource, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	var doc models.DummyResource
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDummyNotFound
		}
		return nil, err
	}

	return &doc, nil
}

func (r *dummyRepositoryImpl) List(ctx context.Context) ([]models.DummyResource, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []models.DummyResource
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []models.DummyResource{}
	}
	return docs, nil
}
