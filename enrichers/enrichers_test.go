package enrichers

import "testing"

func TestDetectServerType(t *testing.T) {
	tests := []struct {
		input      string
		serverType ServerType
	}{
		{"http://google.com", HTTP},
		{"http://1.2.3.4:9200", ELK},
		{"ftp://1.2.3.4:21", FTP},
		{"root:pass@tcp(127.0.0.1:3306)/test", SQL},
	}

	for _, test := range tests {
		if got := DetectServerType(test.input); got != test.serverType {
			t.Errorf("Error on test \"%s\"", test.input)
		}
	}
}
