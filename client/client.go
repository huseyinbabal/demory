package main

import (
	"context"
	"fmt"
	"github.com/huseyinbabal/demory/proto"
	"google.golang.org/grpc"
	"log"
)

func main() {
	conn, connErr := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if connErr!=nil{
		log.Fatalf("connection error %v",connErr)
	}
	defer conn.Close()

	c:= proto.NewDemoryClient(conn)
	_, putErr := c.Put(context.Background(), &proto.PutRequest{
		Key:   "hello",
		Value: "world",
	})
	if putErr != nil {
		log.Fatalf("put err %v",putErr)
	}

	value, valueErr := c.Get(context.Background(), &proto.GetRequest{Key: "hello1"})
	if valueErr != nil {
		log.Fatalf("get err %v",valueErr)
	}
	fmt.Println(value)
}
