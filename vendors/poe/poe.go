package poe

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	"github.com/danielmiessler/fabric/common"
)

// TODO: later refactor all the below utils to a different package

type PoeClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewPoeClient() *PoeClient {
	return &PoeClient{
		BaseURL:    "http://127.0.0.1:8000",
		HTTPClient: &http.Client{},
	}
}

func (p *PoeClient) ListModels(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/ListModels", p.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ListModels request: %w", err)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send ListModels request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK ListModels response: %w", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ListModels response body: %w", err)
	}

	var models []string
	err = json.Unmarshal(body, &models)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ListModels response: %w", err)
	}

	return models, nil
}

type Message struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

// SendRequest represents the request payload for sending a message
type SendRequest struct {
	Model    string     `json:"Model"`
	Messages []*Message `json:"Messages"`
}

// SendResponse represents the response payload from the server
type SendResponse struct {
	Content string `json:"Content"`
}

// SendMessage sends a message to the server and returns the response content
func (p *PoeClient) SendMessages(ctx context.Context, api_key string, model string, messages []*Message) (string, error) {
	url := fmt.Sprintf("%s/Send", p.BaseURL)

	// Create the request payload
	payload := SendRequest{
		Model:    model,
		Messages: messages,
	}

	// Serialize the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a new HTTP POST request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create Send request: %w", err)
	}

	// Set headers
	req.Header.Add("Content-Type", "application/json")
	if api_key != "" {
		req.Header.Add("Api-Key", api_key)
	}

	// Send the request
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send Send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code is OK
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK Send response: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Send response body: %w", err)
	}

	// Parse the JSON response
	var sendResponse SendResponse
	err = json.Unmarshal(body, &sendResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse Send response: %w", err)
	}

	return sendResponse.Content, nil
}

// SendStream sends a message to the server and parses the event stream response
func (p *PoeClient) SendStream(ctx context.Context, api_key string, model string, messages []*Message, channel chan string) error {
	url := fmt.Sprintf("%s/SendStream", p.BaseURL)

	// Create the request payload
	payload := SendRequest{
		Model:    model,
		Messages: messages,
	}

	// Serialize the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a new HTTP POST request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Add("Content-Type", "application/json")
	if api_key != "" {
		req.Header.Add("Api-Key", api_key)
	}

	// Send the request
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SendStream request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code is OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK SendStream response: %d", resp.StatusCode)
	}

	// Read the response as a stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Process each line of the event stream
		channel <- line
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SendStream response stream: %w", err)
	}

	close(channel)
	return nil
}

// END TODO

type Client struct {
	*common.Configurable
	ApiKey    *common.SetupQuestion
	ApiClient *PoeClient
}

func NewClient() (ret *Client) {
	vendorName := "Poe"
	ret = &Client{}

	ret.Configurable = &common.Configurable{
		Label:         vendorName,
		EnvNamePrefix: common.BuildEnvVariablePrefix(vendorName),
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

func toMessages(msgs []*common.Message) (messages []*Message) {
	for _, message := range msgs {
		messages = append(messages, &Message{Role: message.Role, Content: message.Content})
	}
	return
}

func (o *Client) Send(msgs []*common.Message, opts *common.ChatOptions) (ret string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ret, err = o.ApiClient.SendMessages(ctx, o.ApiKey.Value, opts.Model, toMessages(msgs))
	return
}

func (o *Client) SendStream(msgs []*common.Message, opts *common.ChatOptions, channel chan string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = o.ApiClient.SendStream(ctx, o.ApiKey.Value, opts.Model, toMessages(msgs), channel)
	if err != nil {
		return fmt.Errorf("failed to send stream: %w", err)
	}
	return
}
