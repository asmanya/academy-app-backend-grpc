package mongodb

import (
	"context"
	"errors"
	"fmt"
	"grpcapi/internals/models"
	"grpcapi/pkg/utils"

	pb "grpcapi/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddTeachersToDb(ctx context.Context, teachersFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal server error")
	}
	defer client.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, len(teachersFromReq))
	for i, pbTeacher := range teachersFromReq {
		newTeachers[i] = mapPbTeacherToModelTeacher(pbTeacher)
	}

	var addedTeachers []*pb.Teacher
	for _, teacher := range newTeachers {
		result, err := client.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value to database")
		}

		objectId, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectId.Hex()
		}

		pbTeacher := mapModelTeacherToPb(*teacher)
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
}

func GetTeachersFromDb(ctx context.Context, sortOptions primitive.D, filter primitive.M) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal server error")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("teachers")

	var cursor *mongo.Cursor
	if len(sortOptions) < 1 {
		cursor, err = coll.Find(ctx, filter)
	} else {
		cursor, err = coll.Find(ctx, filter, options.Find().SetSort(sortOptions))
	}

	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer cursor.Close(ctx)

	teachers, err := decodeEntities(ctx, cursor, func() *pb.Teacher { return &pb.Teacher{} }, newModel)
	if err != nil {
		return nil, err
	}
	return teachers, nil
}

func newModel() *models.Teacher { return &models.Teacher{} }

func ModifyTeachersInDb(ctx context.Context, pbTeachers []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal server error")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher
	for _, teacher := range pbTeachers {
		if teacher.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id cannot be blank"), "id cannot be blank")
		}

		modelTeacher := mapPbTeacherToModelTeacher(teacher)

		objId, err := primitive.ObjectIDFromHex(teacher.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "invalid id")
		}

		// Convert modelTeacher to BSON document
		modelDoc, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal server error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal server error")
		}

		// Remove the _id field from the updated document
		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("teachers").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintln("error updating teacher id:", teacher.Id))
		}

		updatedTeacher := mapModelTeacherToPb(*modelTeacher)
		updatedTeachers = append(updatedTeachers, updatedTeacher)
	}
	return updatedTeachers, nil
}
