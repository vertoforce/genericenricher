package enrichers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testingServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Data"))
	}))
	return server
}

func TestNewHTTP(t *testing.T) {
	server := testingServer()
	defer server.Close()

	http, err := NewHTTP(server.URL)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Read all data
	body, err := ioutil.ReadAll(http)

	// Check for existence of key parts
	if strings.Index(string(body), "Data") == -1 {
		t.Errorf("Did not read data correctly")
	}
	if strings.Index(string(body), "Date:") == -1 {
		t.Errorf("Did not read data correctly")
	}
	if strings.Index(string(body), "Content-Length:") == -1 {
		fmt.Println(string(body))
		t.Errorf("Did not read data correctly")
	}
}
