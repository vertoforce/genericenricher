package enrichers

import (
	"fmt"
	"net/url"
	"regexp"
)

// ServerType Type of server (ELK, FTP, etc)
type ServerType int

// Server Types
const (
	Unknown ServerType = iota
	ELK
	FTP
	SSH
	SQL
	HTTP
)

type serverRegex struct {
	serverType ServerType
	regex      *regexp.Regexp
}

// serverTypeRegexes array of regexes to determine server type.  Array because order matters to check regexes
var serverTypeRegexes = []serverRegex{
	{ELK, regexp.MustCompile("https?://.*9200")},
	{FTP, regexp.MustCompile("ftp")},
	{SQL, regexp.MustCompile(`(\w*):?(\w*)@(tcp|udp)`)}, // TODO: Finish regex

	// Basic regexes
	{HTTP, regexp.MustCompile("https?://.*")},
}

// DetectServerType Get type of server by looking at URL and/or poking at server
func DetectServerType(connectString string) ServerType {
	connectURL, err := url.Parse(connectString)
	if err != nil {
		return Unknown
	}
	fmt.Println(connectURL)

	// Check regexes
	for _, serverTypeRegex := range serverTypeRegexes {
		if serverTypeRegex.regex.MatchString(connectString) {
			return serverTypeRegex.serverType
		}
	}

	// TODO: Check if multiple matched
	// TODO: Attempt connection to server if there is no protocol

	return Unknown
}
