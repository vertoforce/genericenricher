package genericenricher

import (
	"context"
	"fmt"

	"github.com/vertoforce/genericenricher/enrichers"
)

func ExampleGetServerWithType() {
	// This code does not check for errors
	server, _ := GetServerWithType("http://localhost:9200", enrichers.ELK)
	_ = server.Connect(context.Background())

	p := make([]byte, 10)
	read, _ := server.Read(p)

	fmt.Printf("Read %d bytes: %v\n", read, p)
}
