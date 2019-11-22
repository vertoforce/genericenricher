package enrichers

import (
	"context"
	"fmt"
)

func ExampleNewELK() {
	_, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		panic(err)
	}
}

func ExampleELKClient_GetIndices() {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		panic(err)
	}

	indices, err := con.GetIndices(context.Background())
	if err != nil {
		panic(err)
	}

	for _, index := range indices {
		fmt.Println(index.Index)
	}
}
