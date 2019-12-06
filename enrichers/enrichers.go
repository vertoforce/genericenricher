package enrichers

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
)

//go:generate stringer -type=ServerType

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

// GetConnectionString Given ip, port, and type get the connection string for a server
func GetConnectionString(ip net.IP, port int, serverType ServerType) string {
	// TODO: Add user/pass?
	switch serverType {
	case ELK:
		return fmt.Sprintf("http://%s:%d", ip.String(), port)
	case FTP:
		return fmt.Sprintf("ftp://%s:%d", ip.String(), port)
	case HTTP:
		return fmt.Sprintf("ssh://%s:%d", ip.String(), port)
	case SQL:
		// TODO:
		return fmt.Sprintf("%s:%d", ip.String(), port)
	default:
		return fmt.Sprintf("%s:%d", ip.String(), port)
	}
}
