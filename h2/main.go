package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"sync"
	"time"

	api "github.com/iredelmeier/grpc-http-shared-port"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		// Looks like this throws an error on shutdown if you
		// don't keep the server running (i.e., without the wait group waiting),
		// but this is a proof-of-concept so ¯\_(ツ)_/¯
		if err := server.ServeTLS(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	if err := server.BlockUntilRunning(ServerStartTimeout); err != nil {
		log.Fatalf("Failed to run server: %s", err)
	}

	log.Printf("Server running at https://%s", api.Address)

	log.Print("Testing unverified TLS connection")
	testConn(grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})))
	log.Print("Success!")

	log.Print("Testing secure TLS connection")
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(api.Certificate); !ok {
		log.Fatal("Failed to add certificate")
	}
	testConn(grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	})))
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
