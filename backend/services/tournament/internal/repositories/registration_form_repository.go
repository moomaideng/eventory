package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrFormNotFound          = errors.New("registration form not found")
	ErrResponseNotFound      = errors.New("registration form response not found")
	ErrResponseAlreadyExists = errors.New("registration form response already submitted")
)

const (
	tournamentFormsCollection     = "tournament_forms"
	tournamentResponsesCollection = "tournament_form_responses"
)

type RegistrationFormRepository interface {
	EnsureIndexes(ctx context.Context) error
	GetFormByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*models.RegistrationForm, error)
	UpsertForm(ctx context.Context, form *models.RegistrationForm) (*models.RegistrationForm, error)
	CreateResponse(ctx context.Context, resp *models.FormResponse) (*models.FormResponse, error)
	UpsertResponse(ctx context.Context, resp *models.FormResponse) (*models.FormResponse, error)
	GetResponseByAccount(ctx context.Context, tournamentID uuid.UUID, accountID uuid.UUID) (*models.FormResponse, error)
	ListResponsesByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]models.FormResponse, error)
	GetResponsesByAccounts(ctx context.Context, tournamentID uuid.UUID, accountIDs []uuid.UUID) ([]models.FormResponse, error)
}

type registrationFormRepositoryImpl struct {
	formsCol     *mongo.Collection
	responsesCol *mongo.Collection
}

func NewRegistrationFormRepository(db *mongo.Database) RegistrationFormRepository {
	return &registrationFormRepositoryImpl{
		formsCol:     db.Collection(tournamentFormsCollection),
		responsesCol: db.Collection(tournamentResponsesCollection),
	}
}

func (r *registrationFormRepositoryImpl) EnsureIndexes(ctx context.Context) error {
	_, err := r.formsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tournament_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = r.responsesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "tournament_id", Value: 1},
			{Key: "account_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *registrationFormRepositoryImpl) GetFormByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*models.RegistrationForm, error) {
	var form models.RegistrationForm
	err := r.formsCol.FindOne(ctx, bson.M{"tournament_id": tournamentID}).Decode(&form)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	return &form, nil
}

func (r *registrationFormRepositoryImpl) UpsertForm(ctx context.Context, form *models.RegistrationForm) (*models.RegistrationForm, error) {
	now := time.Now().UTC()
	filter := bson.M{"tournament_id": form.TournamentID}

	update := bson.M{
		"$set": bson.M{
			"title":       form.Title,
			"description": form.Description,
			"fields":      form.Fields,
			"updated_at":  now,
		},
		"$setOnInsert": bson.M{
			"tournament_id": form.TournamentID,
			"created_at":    now,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.formsCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return r.GetFormByTournamentID(ctx, form.TournamentID)
}

func (r *registrationFormRepositoryImpl) CreateResponse(ctx context.Context, resp *models.FormResponse) (*models.FormResponse, error) {
	res, err := r.responsesCol.InsertOne(ctx, resp)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrResponseAlreadyExists
		}
		return nil, err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		resp.ID = oid
	}
	return resp, nil
}

func (r *registrationFormRepositoryImpl) UpsertResponse(ctx context.Context, resp *models.FormResponse) (*models.FormResponse, error) {
	filter := bson.M{
		"tournament_id": resp.TournamentID,
		"account_id":    resp.AccountID,
	}
	update := bson.M{
		"$set": bson.M{
			"answers": resp.Answers,
		},
		"$setOnInsert": bson.M{
			"tournament_id": resp.TournamentID,
			"account_id":    resp.AccountID,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.responsesCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return r.GetResponseByAccount(ctx, resp.TournamentID, resp.AccountID)
}

func (r *registrationFormRepositoryImpl) GetResponseByAccount(ctx context.Context, tournamentID uuid.UUID, accountID uuid.UUID) (*models.FormResponse, error) {
	var resp models.FormResponse
	err := r.responsesCol.FindOne(ctx, bson.M{
		"tournament_id": tournamentID,
		"account_id":    accountID,
	}).Decode(&resp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrResponseNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (r *registrationFormRepositoryImpl) ListResponsesByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]models.FormResponse, error) {
	cursor, err := r.responsesCol.Find(ctx, bson.M{"tournament_id": tournamentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.FormResponse
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.FormResponse{}
	}
	return list, nil
}

func (r *registrationFormRepositoryImpl) GetResponsesByAccounts(ctx context.Context, tournamentID uuid.UUID, accountIDs []uuid.UUID) ([]models.FormResponse, error) {
	responses := make([]models.FormResponse, 0)
	if len(accountIDs) == 0 {
		return responses, nil
	}
	cursor, err := r.responsesCol.Find(ctx, bson.M{"tournament_id": tournamentID, "account_id": bson.M{"$in": accountIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &responses); err != nil {
		return nil, err
	}
	return responses, nil
}
