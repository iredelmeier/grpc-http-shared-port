package main

import (
	"context"
	"log"
	"sync"
	"time"

	api "github.com/iredelmeier/grpc-http-shared-port"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	ServerStartTimeout = time.Second
)

func main() {
	server := api.NewServer()
	defer server.Shutdown(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	if err := server.BlockUntilRunning(ServerStartTimeout); err != nil {
		log.Fatalf("Failed to run server: %s", err)
	}

	log.Printf("Server running at http://%s", api.Address)

	log.Print("Testing insecure connection")
	testConn(grpc.WithInsecure())
	log.Print("Success!")

	wg.Wait()
}

func testConn(opts ...grpc.DialOption) {
	conn, err := grpc.Dial(api.Address, opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %s", err)
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)
	res, err := client.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "world",
	})
	if err != nil {
		log.Printf("Failed to greet: %s", err)
	} else {
		log.Printf("Greeting: %s", res.Message)
	}
}
