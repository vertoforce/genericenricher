package enrichers

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
)

// HTTPClient HTTP Client
type HTTPClient struct {
	url          *url.URL
	reader       io.ReadCloser
	response     *http.Response
	readerCtx    context.Context
	readerCancel context.CancelFunc
}

// NewHTTP Create new HTTP client
func NewHTTP(urlString string) (*HTTPClient, error) {
	client := &HTTPClient{}

	// Parse URL
	var err error
	client.url, err = url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Connect and open reader
func (client *HTTPClient) Connect(ctx context.Context) error {
	req, err := http.NewRequest("GET", client.url.String(), nil)
	req = req.WithContext(ctx)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	client.response = response
	return nil
}

// GetIP Get IP
func (client *HTTPClient) GetIP() net.IP {
	return urlToIP(client.url)
}

// GetPort Get port
func (client *HTTPClient) GetPort() uint16 {
	return urlToPort(client.url)
}

// GetConnectString Get connect string
func (client *HTTPClient) GetConnectString() string {
	return client.url.String()
}

// IsConnected Is server connected.  Will attempt to open a connection
func (client *HTTPClient) IsConnected() bool {
	return client.reader != nil
}

// Type Returns HTTP
func (client *HTTPClient) Type() ServerType {
	return HTTP
}

// Close the connection
func (client *HTTPClient) Close() error {
	// Stop current reader
	if client.reader != nil {
		client.readerCancel()
		err := client.reader.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Read headers, cookies, then body
func (client *HTTPClient) Read(p []byte) (n int, err error) {
	if client.reader == nil {
		err = client.ResetReader()
		if err != nil {
			return 0, err
		}
	}

	return client.reader.Read(p)
}

// ResetReader reset reader back to initial state
func (client *HTTPClient) ResetReader() error {
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
	reader, err := client.Dump(client.readerCtx)
	if err != nil {
		return err
	}

	client.reader = reader

	return nil
}

// -- HTTP specific functions ---

type httpResponse struct {
	response *http.Response
}

// Dump headers, cookies, then body
func (client *HTTPClient) Dump(ctx context.Context) (io.ReadCloser, error) {
	dumpReader, dumpWriter := io.Pipe()

	go func() {
		defer dumpWriter.Close()

		// TODO: stop on cancel

		// Write headers
		for name, values := range client.response.Header {
			dumpWriter.Write([]byte(name))
			dumpWriter.Write([]byte(":"))
			for i, value := range values {
				dumpWriter.Write([]byte(value))
				if i != len(values)-1 {
					dumpWriter.Write([]byte(","))
				}
			}
		}

		// Write cookies
		for _, cookie := range client.response.Cookies() {
			dumpWriter.Write([]byte(cookie.Name))
			dumpWriter.Write([]byte(":"))
			dumpWriter.Write([]byte(cookie.Value))
		}

		// write body
		io.Copy(dumpWriter, client.response.Body)
	}()

	return dumpReader, nil
}
