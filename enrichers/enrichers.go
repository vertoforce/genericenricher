package enrichers

// ServerType Type of server (ELK, FTP, etc)
type ServerType int

// Server Types
const (
	Unknown ServerType = iota
	ELK
	FTP
	SSH
)

// DetectServerType Get type of server by looking at URL and/or poking at server
func DetectServerType(connectString string) ServerType {
	// TODO: implement
	return Unknown
}
