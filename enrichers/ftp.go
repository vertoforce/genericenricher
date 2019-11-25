package enrichers

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"path"
	"regexmachine"
	"regexp"
	"strconv"
	"time"

	"github.com/vertoforce/genericenricher/enrichers/readerhelp"

	"github.com/jlaffaye/ftp"
)

const (
	maxDepth = 100
)

// FTPClient abstracted FTP client
type FTPClient struct {
	username    string
	password    string
	url         *url.URL
	client      *ftp.ServerConn
	readerState *readerhelp.ReaderState
}

// NewFTP Connect to FTP server with provided credentials
func NewFTP(urlString string) (*FTPClient, error) {
	client := &FTPClient{}
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	client.url = url

	// Get user/pass
	if client.url.User != nil {
		// We have creds here
		client.username = client.url.User.Username()
		if password, passwordSet := client.url.User.Password(); passwordSet {
			client.password = password
		}
	}

	// Connect
	err = client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Connect to FTP server
func (client *FTPClient) Connect() error {
	c, err := ftp.Dial(net.JoinHostPort(client.url.Hostname(), client.url.Port()), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}

	// Login
	err = c.Login(client.username, client.password)
	if err != nil {
		return err
	}
	client.client = c

	return nil
}

// GetIP Get IP of server
func (client *FTPClient) GetIP() net.IP {
	addrs, err := net.LookupHost(client.url.Hostname())
	if err != nil || len(addrs) == 0 {
		return net.IP{}
	}
	return net.ParseIP(addrs[0])
}

// GetPort Get Port of server
func (client *FTPClient) GetPort() uint16 {
	port, err := strconv.ParseInt(client.url.Port(), 10, 16)
	if err != nil {
		return 0
	}
	return uint16(port)
}

// IsConnected Is server connected
func (client *FTPClient) IsConnected() bool {
	if client.client == nil {
		return false
	}

	return client.client.NoOp() == nil
}

// Type Returns FTP
func (client *FTPClient) Type() ServerType {
	return FTP
}

// Close connection
func (client *FTPClient) Close() error {
	// Stop current reader
	if client.readerState != nil {
		client.readerState.Stop()
	}

	return client.client.Quit()
}

func (client *FTPClient) Read(p []byte) (n int, err error) {
	if client.readerState == nil {
		client.ResetReader()
	}

	return client.readerState.Read(p)
}

// ResetReader reader back to initial state
func (client *FTPClient) ResetReader() error {
	// Stop current reader
	if client.readerState != nil {
		client.readerState.Stop()
	}

	// Start new reader
	client.readerState = readerhelp.New(context.Background())
	entries := client.getAllData(client.readerState.ReadCtx)
	client.readerState.SetEntries(entries)

	return nil
}

// getAllData Reads all files on server.  Opens a new connection
func (client *FTPClient) getAllData(ctx context.Context) chan []byte {
	fileDatas := make(chan []byte)

	// Make new connection as to not overlap with the master connection
	ourClient, err := NewFTP(client.url.String())
	if err != nil {
		close(fileDatas)
		return fileDatas
	}

	go func() {
		defer close(fileDatas)
		defer ourClient.Close()

		files, err := ourClient.GetAllFilesInFolder(ctx, ".")
		if err != nil {
			return
		}

		for file := range files {
			fileResp, err := ourClient.client.Retr(file)
			if err != nil {
				continue
			}
			// Read to []byte
			// TODO: Use reader?
			fileData, err := ioutil.ReadAll(fileResp)
			fileResp.Close()
			if err != nil {
				continue
			}

			select {
			case fileDatas <- fileData:
			case <-ctx.Done():
				return
			}
		}
	}()

	return fileDatas
}

// GetAllFilesInFolder Get all file paths in FTP folder
func (client *FTPClient) GetAllFilesInFolder(ctx context.Context, dir string) (chan string, error) {
	files := make(chan string)

	if !client.IsConnected() {
		return nil, errors.New("not connected")
	}

	// Get files
	go func() {
		defer close(files)

		// Get entries in this folder
		entries, err := client.client.List(dir)
		if err != nil {
			return
		}

		// Iterate over
		for _, entry := range entries {
			if entry.Type == ftp.EntryTypeFolder {
				if entry.Name != "." && entry.Name != ".." {
					// This is a directory, go recursive
					filesSub, err := client.GetAllFilesInFolder(ctx, entry.Name)
					if err != nil {
						return
					}
					for file := range filesSub {
						select {
						case files <- entry.Name + "/" + file:
						case <-ctx.Done():
							return
						}
					}
				}
			} else if entry.Type == ftp.EntryTypeFile {
				// This is a file
				select {
				case files <- entry.Name:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return files, nil
}

// GetFilesMatchingRules Get files that have content that matches the rules
func (client *FTPClient) GetFilesMatchingRules(ctx context.Context, rules []*regexp.Regexp, maxFileDownloadSize int64, maxFilesToCheck int64) ([]string, error) {
	// Recursively fetch all documents
	matchedFiles, err := client.GetFilesMatchingRulesInDir(ctx, ".", rules, maxFileDownloadSize, maxFilesToCheck)
	if err != nil {
		return nil, err
	}

	return matchedFiles, nil
}

// GetFilesMatchingRulesInDir Get files that have content that matches the rules in the folder
func (client *FTPClient) GetFilesMatchingRulesInDir(ctx context.Context, dir string, rules []*regexp.Regexp, maxFileDownloadSize, maxFilesToCheck int64) (matchedFiles []string, err error) {
	checkedFiles := int64(0)
	return client.getFilesMatchingRulesInDirInner(ctx, dir, &checkedFiles, rules, maxFileDownloadSize, maxFilesToCheck, 0)
}

func (client *FTPClient) getFilesMatchingRulesInDirInner(ctx context.Context, dir string, checkedFiles *int64, rules []*regexp.Regexp, maxFileDownloadSize, maxFilesToCheck int64, depth int64) ([]string, error) {
	// Ditch if we are too far down
	if depth >= maxDepth {
		return nil, errors.New("max depth exceeded")
	}

	matchedFiles := []string{}

	// Check contents of each file
	files, err := client.GetAllFilesInFolder(ctx, dir)
	if err != nil {
		return nil, err
	}

	for file := range files {
		fileData, err := client.client.Retr(path.Join(dir, file))
		if err != nil {
			continue
		}

		// Read data
		readCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		matchedRules := regexmachine.RuleSet(rules).GetMatchedRulesReader(readCtx, ioutil.NopCloser(io.LimitReader(fileData, maxFileDownloadSize)))
		cancel()
		fileData.Close()

		if len(matchedRules) > 0 {
			matchedFiles = append(matchedFiles, file)
		}

		(*checkedFiles)++
		// Check if we already read enough files
		if *checkedFiles >= maxFilesToCheck {
			break
		}
	}

	return matchedFiles, nil
}
