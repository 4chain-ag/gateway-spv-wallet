package gateway

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type StablecoinFee struct {
	CommissionRecipient string  `json:"commissionRecipient"`
	From                uint64  `json:"from"`
	To                  uint64  `json:"to"`
	Type                string  `json:"type"`
	Value               float64 `json:"value"`
}

type StablecoinRule struct {
	CoinSym string           `json:"coinSym"`
	TokenId string           `json:"tokenId"`
	Fees    []*StablecoinFee `json:"fees"`
}

type Client interface {
	GetStablecoinRules(tokenID string) (*StablecoinRule, error)
}

type gatewayClient struct {
	log        zerolog.Logger
	gatewayURL string
	httpClient *resty.Client
}

func NewGatewayClient(logger *zerolog.Logger, gatewayURL string, httpClient *resty.Client) (Client, error) {
	return &gatewayClient{
		log:        logger.With().Str("tokens", "gateway-client").Logger(),
		gatewayURL: gatewayURL,
		httpClient: httpClient,
	}, nil
}

func (c *gatewayClient) GetStablecoinRules(tokenID string) (*StablecoinRule, error) {
	url := fmt.Sprintf("%s/coins/bsv21/rules?tokenId=%s", c.gatewayURL, tokenID)
	var response StablecoinRule
	_, err := c.httpClient.R().
		SetResult(&response).
		Get(url)
	if err != nil {
		return nil, err
	}

	result := &StablecoinRule{
		CoinSym: response.CoinSym,
		TokenId: response.TokenId,
	}

	for _, r := range response.Fees {
		if r.CommissionRecipient != "" {
			result.Fees = append(result.Fees, r)
		}
	}

	return result, nil
}
