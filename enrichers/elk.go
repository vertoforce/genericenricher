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

	"github.com/olivere/elastic"
)

// ELKClient ELK Connection
type ELKClient struct {
	url          *url.URL
	client       *elastic.Client
	reader       io.ReadCloser
	readerCtx    context.Context
	readerCancel context.CancelFunc
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
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	client.url = url

	// Connect
	err = client.Connect()
	if err != nil {
		return nil, err
	}

	// TODO: Check for user/pass

	return &client, nil
}

// Connect to ELK server
func (client *ELKClient) Connect() error {
	var err error
	client.client, err = elastic.NewSimpleClient(elastic.SetURL(client.url.String()))
	if err != nil {
		return err
	}

	return nil
}

// GetIP Get IP of server
func (client *ELKClient) GetIP() net.IP {
	return urlToIP(client.url)
}

// GetPort Get Port of server
func (client *ELKClient) GetPort() uint16 {
	return urlToPort(client.url)
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
	// Stop current reader
	if client.reader != nil {
		client.readerCancel()
		err := client.reader.Close()
		if err != nil {
			return err
		}
	}

	client.client.Stop()
	return nil
}

// Read Returns all data from all indices on server
func (client *ELKClient) Read(p []byte) (n int, err error) {
	if !client.IsConnected() {
		return 0, errors.New("not connected")
	}
	if client.reader == nil {
		err = client.ResetReader()
		if err != nil {
			return 0, err
		}
	}

	return client.reader.Read(p)
}

// ResetReader reset reader back to initial state
func (client *ELKClient) ResetReader() error {
	// Stop current reader
	if client.reader != nil {
		client.readerCancel()
		err := client.reader.Close()
		if err != nil {
			return err
		}
	}

	// Start new reader
	client.readerCtx, client.readerCancel = context.WithCancel(context.Background())
	client.reader = client.getAllData(client.readerCtx)

	return nil
}

// -- ELK specific functions --

// getAllData Get all entries in all indices
func (client *ELKClient) getAllData(ctx context.Context) io.ReadCloser {
	allDataReader, allDataWriter := io.Pipe()

	// Make sure we are connected
	if !client.IsConnected() {
		allDataWriter.Close()
		return allDataReader
	}

	go func() {
		defer allDataWriter.Close()

		// Go through every index
		indices, err := client.GetIndices(ctx)
		if err != nil {
			return
		}
		for _, index := range indices {
			// For every entry in this index
			indexEntries := client.GetJSONData(ctx, index.Index, -1)
			for entry := range indexEntries {
				// TODO: Cancel with context
				allDataWriter.Write(*entry)
			}
		}
		allDataWriter.Close()
	}()

	return allDataReader
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
			stringSizeToUint(index.StoreSize),
		})
	}

	return ret, nil
}

// GetTotalSize Gets total size of the ELK instance in bytes by summing all the sizes of each index
func (client *ELKClient) GetTotalSize(ctx context.Context) (uint64, error) {
	indices, err := client.GetIndices(ctx)
	if err != nil {
		return 0, err
	}

	totalSize := uint64(0)
	for _, index := range indices {
		totalSize += index.StoreSize
	}

	return totalSize, nil
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
