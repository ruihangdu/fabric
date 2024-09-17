package poe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/danielmiessler/fabric/common"
)

// TODO: later refactor all the below utils to a different package

type PoeClient struct {
	BaseURL 	string
	HTTPClient 	*http.Client
}

func NewPoeClient() *PoeClient {
	return &PoeClient{
		BaseURL: "http://127.0.0.1:8000",
		HTTPClient: &http.Client{},
	}
}

func (p *PoeClient) ListModels(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/ListModels", p.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ListModels request: %w", err)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send ListModels request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received non-OK ListModels response: %w", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read ListModels response body: %w", err)
	}

	var models []string
	err = json.Unmarshal(body, &models)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse ListModels response: %w", err)
	}

	return models, nil
}
// END TODO

type Client struct {
	*common.Configurable
	ApiKey *common.SetupQuestion
	ApiClient *PoeClient
}

func NewClient() (ret *Client) {
	vendorName := "Poe"
	ret = &Client{}

	ret.Configurable = &common.Configurable{
		Label:			vendorName,
		EnvNamePrefix:	common.BuildEnvVariablePrefix(vendorName),
	}

	ret.ApiKey = ret.Configurable.AddSetupQuestion("API key", true)
	ret.ApiClient = NewPoeClient()

	return
}

func (o *Client) ListModels() (ret []string, err error) {
	// Currently a hardcoded list of models because Poe doesn't have API to get all the "bots"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ret, err = o.ApiClient.ListModels(ctx)
	return
}

func (o *Client) Send(msgs []*common.Message, opts *common.ChatOptions) (ret string, err error) {
	return
}

func (o *Client) SendStream(msgs []*common.Message, opts *common.ChatOptions, channel chan string) (err error) {
	return
}
