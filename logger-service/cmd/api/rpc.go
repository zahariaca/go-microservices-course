package main

import (
	"context"
	"github.com/zahariaca/logger-service/data"
	"log"
	"time"
)

// RPCServer is the type of our RPC server, Methods that take this as a receiver are available
// over RPC, as long as they are exported.
type RPCServer struct {
}

// RPCPayload is the type for data we receive from RPC
type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	log.Println("Entered RPCServer.LogInfo with payload:", payload)
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("error writing to mongo", err)
		return err
	}

	*resp = "Processed payload via RPC:" + payload.Name
	return nil
}
