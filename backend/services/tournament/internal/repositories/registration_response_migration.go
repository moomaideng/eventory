package repositories

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// migrateResponseKeys consolidates the former per-membership answer history.
// ObjectIDs order old inserts chronologically; keep the latest answer per key.
func (r *registrationFormRepositoryImpl) migrateResponseKeys(ctx context.Context) error {
	const oldIndex = "tournament_id_1_account_id_1_registration_id_1"
	indexes, err := r.responsesCol.Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer indexes.Close(ctx)
	var existing []struct {
		Name string `bson:"name"`
	}
	if err := indexes.All(ctx, &existing); err != nil {
		return err
	}
	legacy := false
	for _, index := range existing {
		legacy = legacy || index.Name == oldIndex
	}
	if !legacy {
		return nil
	}
	cursor, err := r.responsesCol.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"tournament_id": "$tournament_id", "account_id": "$account_id"},
			"ids": bson.M{"$push": "$_id"},
		}}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var group struct {
			IDs []bson.ObjectID `bson:"ids"`
		}
		if err := cursor.Decode(&group); err != nil {
			return err
		}
		if len(group.IDs) > 1 {
			if _, err := r.responsesCol.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": group.IDs[1:]}}); err != nil {
				return err
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if _, err := r.responsesCol.UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"team_id": "", "registration_id": ""}}); err != nil {
		return err
	}
	err = r.responsesCol.Indexes().DropOne(ctx, oldIndex)
	var commandError mongo.CommandError
	if errors.As(err, &commandError) && commandError.Code == 27 {
		return nil
	}
	return err
}
