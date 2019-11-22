package enrichers

import (
	"context"
	"fmt"
	"io"
	"regexmachine"
	"testing"
)

var (
	serverHost = "ftp://username:mypass@localhost:21"
)

func TestGetFiles(t *testing.T) {
	client, err := NewFTP(serverHost)
	if err != nil {
		t.Errorf("failed to connect")
	}
	defer client.Close()

	// Check if we can get files
	files, err := client.GetAllFilesInFolder(context.Background(), ".")
	fileCount := 0
	for file := range files {
		fmt.Println(file)
		fileCount++
	}
	if fileCount == 0 {
		t.Errorf("No files found")
	}
}

func TestReadFTP(t *testing.T) {
	client, err := NewFTP(serverHost)
	if err != nil {
		t.Errorf("failed to connect")
		return
	}
	defer client.Close()

	p := make([]byte, 1024)
	read, err := client.Read(p)
	fmt.Println(string(p))
	if read == 0 || (err != nil && err != io.EOF) {
		t.Errorf("Error reading")
	}
}

func TestGetFilesMatchingRules(t *testing.T) {
	client, err := NewFTP(serverHost)
	if err != nil {
		t.Errorf("failed to connect")
	}
	defer client.Close()

	files, err := client.GetFilesMatchingRules(context.Background(), regexmachine.MatchAll, 1024*1024*1024, 10)
	if err != nil {
		t.Errorf(err.Error())
	}
	// Make sure we got some files
	if len(files) == 0 {
		t.Errorf("No files found")
	}

	// Check # of files limit
	files, err = client.GetFilesMatchingRules(context.Background(), regexmachine.MatchAll, 1024*1024*1024, 1)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(files) != 1 {
		t.Errorf("Not obeying file limit")
	}
}
