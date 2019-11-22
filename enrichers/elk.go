package enrichers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"regexmachine"
	"regexp"
	"strconv"

	"github.com/olivere/elastic"
)

// ELKClient ELK Connection
type ELKClient struct {
	IP   net.IP
	Port int16
	URL  string

	// Internal
	client *elastic.Client
	// Reading stream
	readCtx          context.Context
	indices          []ELKIndex
	curIndex         int                   // Current Index
	curIndexReader   chan *json.RawMessage // Current Index reader (from GetJSONData)
	curRawMessage    *json.RawMessage      // Current JSON from index
	curRawMessagePos int                   // Current position in JSON from index
	curPPos          int                   // Current position in given buffer
}

// ELKIndex ELK Index
type ELKIndex struct {
	Health             string
	Status             string
	Index              string
	UUID               string
	Pri                int
	Rep                int
	DocsCount          int
	DocsDeleted        int
	CreationDate       int64
	CreationDateString string
	StoreSize          uint64 // Store size in bytes
}

// NewELK Connect to ELK server
func NewELK(urlString string) (*ELKClient, error) {
	client := ELKClient{}
	// Set URL
	client.URL = urlString
	urlObj, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	// Set Port
	port, err := strconv.ParseInt(urlObj.Port(), 10, 16)
	if err != nil {
		return nil, err
	}
	client.Port = int16(port)

	// Set IP
	addrs, err := net.LookupHost(urlObj.Hostname())
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New("Invalid hostname")
	}
	client.IP = net.ParseIP(addrs[0])

	// Set internals
	client.Reset()

	return &client, client.Connect()
}

// Connect to ELK server
func (client *ELKClient) Connect() error {
	var err error
	client.client, err = elastic.NewSimpleClient(elastic.SetURL(client.URL))
	if err != nil {
		return err
	}

	return nil
}

// GetIP Get IP of server
func (client *ELKClient) GetIP() net.IP {
	return client.IP
}

// GetPort Get Port of server
func (client *ELKClient) GetPort() int16 {
	return client.Port
}

// IsConnected Is server connected
func (client *ELKClient) IsConnected() bool {
	return client.client.IsRunning()
}

// Type Returns ELK
func (client *ELKClient) Type() ServerType {
	return ELK
}

// Close server
func (client *ELKClient) Close() error {
	client.curIndex = -1
	client.client.Stop()
	return nil
}

// Read Returns all data from all indices on server
func (client *ELKClient) Read(p []byte) (n int, err error) {
	// Make sure we are connected
	if !client.IsConnected() {
		return 0, errors.New("not connected to server")
	}

	// Check what state we are in
	switch {
	case client.curIndex == -1:
		// Get indices and pick first one
		client.indices, err = client.GetIndices(client.readCtx)
		if err != nil {
			return 0, err
		}
		client.curIndex = 0

		// Recurs to next step
		return client.Read(p)
	case client.curIndexReader == nil:
		// Check if we finished reading indices
		if client.curIndex >= len(client.indices) {
			return 0, io.EOF
		}

		// Open index for reading
		client.curIndexReader = client.GetJSONData(client.readCtx, client.indices[client.curIndex].Index, -1)

		// Recurs to next step
		return client.Read(p)
	case client.curRawMessage == nil:
		// Start reading from this index
		buf, ok := <-client.curIndexReader
		if !ok {
			// end of this index, move on
			client.curIndex++
			client.curIndexReader = nil
			return client.Read(p)
		}
		client.curRawMessage = buf
		client.curRawMessagePos = 0

		// Recurs to next step
		return client.Read(p)
	default:
		// Now we have a buf

		// Copy to p and see if we need to copy more
		copied, needNextSource := readIntoP(p, &client.curPPos, &client.curRawMessagePos, *client.curRawMessage)
		if needNextSource {
			// Need to read more data into this p

			// Reset to go to next entry
			client.curRawMessage = nil

			// Read more data
			read, err := client.Read(p)
			return read + copied, err
		}

		// We are done
		return copied, nil
	}
}

// Reset reader back to initial state
func (client *ELKClient) Reset() {
	client.readCtx = context.Background()
	client.indices = nil
	client.curIndex = -1
	client.curIndexReader = nil
	client.curRawMessage = nil
	client.curRawMessagePos = 0
	client.curPPos = 0
}

// GetIndicesMatchingRules Return all indices that have contents that match a rule in the provided ruleset.
func (client *ELKClient) GetIndicesMatchingRules(ctx context.Context, rules []*regexp.Regexp, maxDocsToCheck int64) ([]ELKIndex, error) {
	// Get indices
	indices, err := client.GetIndices(ctx)
	if err != nil {
		return nil, err
	}

	matchedIndices := []ELKIndex{}

	// Search data of each index
	for _, index := range indices {
		dataCtx, cancel := context.WithCancel(ctx)
		dataStream := client.GetData(dataCtx, index.Index, maxDocsToCheck)

		// Check all docs
		for hit := range dataStream {
			ruleSet := regexmachine.RuleSet(rules)
			if len(ruleSet.GetMatchedRules(*hit.Source)) > 0 {
				matchedIndices = append(matchedIndices, index)
				break
			}
		}

		// Stop fetching data (especially if we broke early)
		cancel()
	}

	return matchedIndices, nil
}

// GetIndices Get indices on server
func (client *ELKClient) GetIndices(ctx context.Context) ([]ELKIndex, error) {
	// Get indices
	indices, err := client.client.CatIndices().Do(ctx)
	if err != nil {
		return nil, err
	}

	ret := []ELKIndex{}

	// Convert to our indices
	for _, index := range indices {
		ret = append(ret, ELKIndex{
			index.Health,
			index.Status,
			index.Index,
			index.UUID,
			index.Pri,
			index.Rep,
			index.DocsCount,
			index.DocsDeleted,
			index.CreationDate,
			index.CreationDateString,
			StringSizeToUint(index.StoreSize),
		})
	}

	return ret, nil
}

// GetJSONData Given index name, return channel of jsons limited to `limit` hits. -1 for unlimited
func (client *ELKClient) GetJSONData(ctx context.Context, indexName string, limit int64) chan *json.RawMessage {
	ret := make(chan *json.RawMessage)

	go func() {
		defer close(ret)

		hits := client.GetData(ctx, indexName, limit)
		for hit := range hits {
			select {
			case ret <- hit.Source:
			case <-ctx.Done(): // Check if canceled
				return
			}
		}
	}()

	return ret
}

// GetData Given an index name, return channel of hits limited to `limit` hits.  -1 for unlimited.
func (client *ELKClient) GetData(ctx context.Context, indexName string, limit int64) chan *elastic.SearchHit {
	ret := make(chan *elastic.SearchHit)

	go func() {
		defer close(ret)

		// Scroll
		scrollService := elastic.NewScrollService(client.client)
		scrollService.Index(indexName)
		scrollService.Size(10)
		defer scrollService.Clear(ctx) // Not sure what context to use here

		// Read all data
		totalHits := int64(0)
		for {
			result, err := scrollService.Do(ctx)
			if err != nil {
				if err == io.EOF {
					return
				}
				// Something went wrong
				return
			}

			// For each doc
			for _, hit := range result.Hits.Hits {
				totalHits++
				select {
				case ret <- hit:
				case <-ctx.Done(): // Check if canceled
					return
				}

				// Check total hits
				if limit != -1 && totalHits >= limit {
					return
				}
			}
		}
	}()

	return ret
}
