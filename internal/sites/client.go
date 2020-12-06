package sites

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Client for doing request to sites
type Client interface {
	GetData(ctx context.Context, url string) (data string, err error)
}

type client struct {
	clientHTTP http.Client
}

func (c *client) GetData(ctx context.Context, url string) (data string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return data, fmt.Errorf("failed to create request: %s", err)
	}

	resp, err := c.clientHTTP.Do(req)
	if err != nil {
		return data, fmt.Errorf("failed to make request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("status code %d for url %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, fmt.Errorf("failed to read body response: %s", err)
	}
	defer resp.Body.Close()

	data = string(body)

	return
}

// NewClient ...
func NewClient(clientHTTP http.Client) Client {
	return &client{
		clientHTTP: clientHTTP,
	}
}
