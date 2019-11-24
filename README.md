# GenericEnricher

The genericenricher package is meant to act as a black box that works like this:

```
ftp://localhost:21 -> genericenricher -> Read() // Read raw data
```

The package was built with the purpose of making it easy to stream data from a generic server right into a regex search.
