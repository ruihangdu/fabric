package poe

import (
	"context"
	"errors"
	"fmt"
    "io/ioutil"
    "net/http"
	"strings"

	"github.com/danielmiessler/fabric/common"
)

type Client struct {
	*common.Configurable
	ApiKey *common.SetupQuestion
}

func NewClient() (ret *Client) {
	vendorName := "Poe"
	ret = &Client{}

	ret.Configurable = &common.Configurable{
		Label:			vendorName,
		EnvNamePrefix:	common.BuildEnvVariablePrefix(vendorName),
	}

	ret.ApiKey = ret.Configurable.AddSetupQuestion("API key", true)

	return
}

func (o *Client) ListModels() (ret []string, err error) {
	// Currently a hardcoded list of models because Poe doesn't have API to get all the "bots"
	return
}

func (o *Client) Send(msgs []*common.Message, opts *common.ChatOptions) (ret string, err error) {
	return
}

func (o *Client) SendStream(msgs []*common.Message, opts *common.ChatOptions, channel chan string) (err error) {
	return
}
