# GenericEnricher

The genericenricher package is meant to act as a black box.
You give a connection url such as `ftp://localhost:21` and are given a set of functions to perform generic actions.
For example:

```go
server, err := GetServer("http://localhost:9200")
if err != nil {
    return
}

p := make([]byte, 10)
server.Read(p)
```

You can also run **specific functions** if you know the server type.  For example:

```go
server, err := GetServer("http://localhost:9200")
if err != nil {
    return
}

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
