package genericenricher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/vertoforce/genericenricher/enrichers"
)

var (
	localhost = net.IP{127, 0, 0, 1}
)

func TestGetServerWithType(t *testing.T) {
	tests := []struct {
		url        string
		serverType enrichers.ServerType
		ip         net.IP
		port       uint16
	}{
		// Local HTTP
		{"http://localhost", enrichers.HTTP, localhost, 80},
		// Local ELK
		{"http://localhost:9200", enrichers.ELK, localhost, 9200},
		// Local FTP
		{"ftp://username:mypass@localhost:21", enrichers.FTP, localhost, 21},
		// Local SQL
		{"mysql://root:pass@tcp(127.0.0.1:3306)/test", enrichers.SQL, localhost, 3306},
	}

	for _, test := range tests {
		server, err := GetServerWithType(test.url, test.serverType)
		if err != nil {
			t.Errorf("Failed to create server")
			continue
		}
		err = checkServerFunctionality(server, test.ip, test.port)
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}

func checkServerFunctionality(s Server, ip net.IP, port uint16) error {
	// Check IP and Port
	if !s.GetIP().Equal(ip) {
		fmt.Println(s.GetIP())
		return errors.New("bad ip")
	}
	if s.GetPort() != port {
		fmt.Println(s.GetPort())
		return errors.New("bad port")
	}

	// Connect
	err := s.Connect(context.Background())
	if err != nil {
		return errors.New("error connecting to server")
	}

	// Check if connected
	if !s.IsConnected() {
		return errors.New("should be connected to server")
	}

	// Check read
	p := make([]byte, 10)
	read, err := s.Read(p)
	if read == 0 {
		return errors.New("Could not read any data")
	}
	if err != nil && err != io.EOF {
		return err
	}

	// Close
	err = s.Close()
	if err != nil {
		return err
	}
	if s.IsConnected() {
		return errors.New("should not be connected to server")
	}

	return nil
}
