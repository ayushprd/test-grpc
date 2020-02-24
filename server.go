package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stetpb "github.com/user/basic-crud/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var collection *mongo.Collection

type server struct {
}

type personItem struct {
	ID   primitive.ObjectID `bson: "_id"`
	Name string             `bson: "name"`
}

func (*server) CreatePerson(ctx context.Context, req *stetpb.CreatePersonRequest) (*stetpb.CreatePersonResponse, error) {
	fmt.Println("Create person request")
	person := req.GetPerson()
	data := personItem{
		Name: person.GetName(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID"),
		)
	}
	return &stetpb.CreatePersonResponse{
		Person: &stetpb.Person{
			Id:   oid.Hex(),
			Name: person.GetName(),
		},
	}, nil
}

func (*server) ReadPerson(ctx context.Context, req *stetpb.ReadPersonRequest) (*stetpb.ReadPersonResponse, error) {
	fmt.Println("Read Person request")
	personID := req.GetPersonId()
	oid, err := primitive.ObjectIDFromHex(personID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}
	data := personItem{}
	filter := bson.M{"_id": oid} //filter := bson.D{{"_id", oid}}
	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find person with specified Id: %v", err),
		)
	}
	response := &stetpb.ReadPersonResponse{
		Person: &stetpb.Person{
			Id:   oid.Hex(),
			Name: data.Name,
		},
	}
	return response, nil
}

/*func dataToPersonPb(data *personItem) *stetpb.Person {
	return &stetpb.Person{
		Id:   data.ID.Hex(),
		Name: data.Name,
	}
}*/

func (*server) UpdatePerson(ctx context.Context, req *stetpb.UpdatePersonRequest) (*stetpb.UpdatePersonResponse, error) {
	fmt.Println("Update person request")
	person := req.GetPerson()
	oid, err := primitive.ObjectIDFromHex(person.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	update := bson.M{
		"name": person.GetName(),
	}
	data := personItem{}
	filter := bson.D{{"_id", oid}}

	res := collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": update})
	if err := res.Decode(&data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find person with specified Id: %v", err),
		)
	}
	return &stetpb.UpdatePersonResponse{
		Person: &stetpb.Person{
			Id:   data.ID.Hex(),
			Name: data.Name,
		},
	}, nil

}

func (*server) DeletePerson(ctx context.Context, req *stetpb.DeletePersonRequest) (*stetpb.DeletePersonResponse, error) {
	fmt.Println("Delete person request")
	oid, err := primitive.ObjectIDFromHex(req.GetPersonId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}
	filter := bson.D{{"_id", oid}}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object in DB: %v", err),
		)
	}
	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find person in DB: %v", err),
		)
	}
	return &stetpb.DeletePersonResponse{PersonId: req.GetPersonId()}, nil
}
func main() {
	//if the code crashes, we get the filename and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")
	//connect to DB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Service started")
	collection = client.Database("mydb").Collection("stet")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	stetpb.RegisterStetServiceServer(s, &server{})
	
	reflection.Register(s)
	
	go func() {
		fmt.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	//CTRL+C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	fmt.Println("End of Program")
}
