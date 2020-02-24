package main

import (
	"context"
	"fmt"
	"log"

	stetpb "github.com/user/basic-crud/proto"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Client")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Could not connect :%v", err)
	}
	defer cc.Close()

	c := stetpb.NewStetServiceClient(cc)

	//create Person
	fmt.Println("Creating the person")
	person := &stetpb.Person{
		Name: "xyz",
	}
	createPersonRes, err := c.CreatePerson(context.Background(), &stetpb.CreatePersonRequest{Person: person})

	if err != nil {
		log.Fatalf("Unexpexted error: %v \n", err)
	}
	fmt.Printf("Person has been created: %v \n", createPersonRes)
	personID := createPersonRes.GetPerson().GetId()

	//read Person
	fmt.Println("Reading the person")

	/*_, err2 := c.ReadPerson(context.Background(), &stetpb.ReadPersonRequest{PersonId: "1dgghd"})
	if err2 != nil {
		fmt.Printf("Error while reading : %v \n", err2)
	}*/
	readPersonReq := &stetpb.ReadPersonRequest{PersonId: personID}
	readPersonRes, readPersonErr := c.ReadPerson(context.Background(), readPersonReq)
	if readPersonErr != nil {
		fmt.Printf("Error while raeading: %v \n", readPersonErr)
	}
	fmt.Printf("Person was read: %v \n", readPersonRes)

	//update Person
	fmt.Println("Updating the person")
	newPerson := &stetpb.Person{
		Id:   personID,
		Name: "Changed xyz",
	}

	updateRes, updateErr := c.UpdatePerson(context.Background(), &stetpb.UpdatePersonRequest{Person: newPerson})

	if updateErr != nil {
		fmt.Printf("Error while updating : %v \n", updateErr)
	}
	fmt.Printf("Person was updated: %v \n", updateRes)

	//delete person
	fmt.Println("Deleting the person")
	deleteRes, deleteErr := c.DeletePerson(context.Background(), &stetpb.DeletePersonRequest{PersonId: personID})
	if deleteErr != nil {
		fmt.Printf("Error while deleting: %v \n", deleteErr)
	}
	fmt.Printf("Person was deleted: %v \n", deleteRes)
}
