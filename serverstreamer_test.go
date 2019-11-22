package serverstreamer

import (
	"errors"
	"io"
	"net"
	"serverstreamer/enrichers"
	"testing"
)

func checkServerFunctionality(s Server, ip net.IP, port uint16) error {
	// Check IP and Port
	if !s.GetIP().Equal(ip) {
		return errors.New("bad ip")
	}
	if s.GetPort() != port {
		return errors.New("bad port")
	}

	// Connect
	err := s.Connect()
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
	if read == 0 || (err != nil && err != io.EOF) {
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

func TestGetServerWithType(t *testing.T) {
	tests := []struct {
		url        string
		serverType enrichers.ServerType
		ip         net.IP
		port       uint16
	}{
		// Local ELK
		{"http://localhost:9200", enrichers.ELK, net.IPv6loopback, 9200},
	}

	for _, test := range tests {
		server, err := GetServerWithType(test.url, test.serverType)
		if err != nil {
			t.Errorf("Failed to create server")
		}
		err = checkServerFunctionality(server, test.ip, test.port)
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}
