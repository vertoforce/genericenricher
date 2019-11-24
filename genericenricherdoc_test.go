package genericenricher

import (
	"fmt"

	"github.com/vertoforce/genericenricher/enrichers"
)

func ExampleGetServerWithType() {
	server, err := GetServerWithType("http://localhost:9200", enrichers.ELK)
	if err != nil {
		return
	}

	err = server.Connect()
	if err != nil {
		return
	}

	p := make([]byte, 10)
	read, err := server.Read(p)
	if err != nil {
		return
	}

	fmt.Printf("Read %d bytes: %v\n", read, p)
}
