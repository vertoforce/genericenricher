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

// DetectServerType Get type of server by poking at server
func DetectServerType(connectString string) ServerType {
	// TODO: implement
	return Unknown
}
