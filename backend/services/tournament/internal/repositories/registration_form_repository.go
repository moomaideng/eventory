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
	GetResponseByAccount(ctx context.Context, tournamentID uuid.UUID, accountID uuid.UUID, registrationID ...uuid.UUID) (*models.FormResponse, error)
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
			{Key: "registration_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	// Build the replacement before dropping the old tournament-wide constraint.
	err = r.responsesCol.Indexes().DropOne(ctx, "tournament_id_1_account_id_1")
	var commandError mongo.CommandError
	if errors.As(err, &commandError) && commandError.Code == 27 {
		return nil
	}
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

func (r *registrationFormRepositoryImpl) GetResponseByAccount(ctx context.Context, tournamentID uuid.UUID, accountID uuid.UUID, registrationID ...uuid.UUID) (*models.FormResponse, error) {
	var resp models.FormResponse
	filter := bson.M{
		"tournament_id": tournamentID,
		"account_id":    accountID,
	}
	if len(registrationID) > 0 {
		filter["registration_id"] = registrationID[0]
	} else {
		filter["registration_id"] = bson.M{"$exists": false}
	}
	err := r.responsesCol.FindOne(ctx, filter).Decode(&resp)
	// Older responses have no membership ID. Keep them read-only for existing
	// memberships; newly created memberships always have their own response.
	if errors.Is(err, mongo.ErrNoDocuments) && len(registrationID) > 0 {
		filter["registration_id"] = bson.M{"$exists": false}
		err = r.responsesCol.FindOne(ctx, filter).Decode(&resp)
	}
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
