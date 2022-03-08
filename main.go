package main

import (
	"encoding/json"
	"fmt"

	"github.com/auth/actions"
	"github.com/auth/db"
)

// Request is the kind of request that is made.
// Ex:
//  Type: ValidateUser, CheckUser, GenerateAuth, DeleteAuth, RefreshAuth.
//  Data: Is the JSON payload that is sent.
//  RequestKey will usually be a uniqie key that the sender will identify the request with - could be the ID of the user
type Request struct {
	Type       string            `json:"type"`
	Data       map[string]string `json:"data"`
	RequestKey string            `json:"request_key"`
}

// ListenForMessages will run for ever listening on a predetermined queue
func ListenForMessages() {
	redis := db.GetRedisConnection()
	subscriber := redis.Connection.Subscribe("auth")

	fmt.Println("---------------------------------")
	fmt.Println("AUTH SERVICE STARTED !!")
	fmt.Println("---------------------------------")
	for {
		msg, err := subscriber.ReceiveMessage()
		if err != nil {
			panic(err)
		}

		request := Request{}

		if err := json.Unmarshal([]byte(msg.Payload), &request); err != nil {
			panic(err)
		}

		fmt.Println("Received message from " + msg.Channel + " channel.")
		fmt.Printf("%+v\n", request)

		if actions.CheckValidRequestType(request.Type) {
			// The following should be called based on requestType
			actions.CheckUserAuth(request.Data)
		}
	}
}

// main function for the auth micro service.
func main() {
	// Get Database connection and setup models to migrate automatically.
	db := db.GetDatabaseConnection()
	db.SetLogger()
	db.MigrateModels()

	// ListenForMessages infinately.
	ListenForMessages()
}
