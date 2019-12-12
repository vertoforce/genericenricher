# GenericEnricher

The genericenricher package is meant to act as a black box.
You give a connection url such as `ftp://localhost:21` and you can read raw data from the server.  You are also given a set of generic functions to call.  The returned server object will be connected to the server.

## Usage

```go
// This code does not check for errors
server, _ := GetServer("http://localhost:9200")
_ = server.Connect(context.Background())

p := make([]byte, 10)
read, _ = server.Read(p)
```

You can also run **specific functions** if you know the server type.  For example:

```go
// This code does not check for errors
server, _ := GetServer("http://localhost:9200")
_ = server.Connect(context.Background())

// Convert to ELK Client
ELKServer := server.(*enrichers.ELKClient)
indices, err := ELKServer.GetIndices(context.Background())

fmt.Println(indices)
```

## Current functions

- Read() // Read raw data from server.  Useful to stream data from a generic server right into a regex search.

## Current supported server types

- FTP (Looking at file data)
- ELK (Looking at data in indices)
- SQL (Reading data in database tables)
- HTTP (Read webpage, might implement crawling in the future)

## Known Issues

- Currently with some early ELK versions the Read() only manages to read the first 100 documents because the server does not support scrolling
