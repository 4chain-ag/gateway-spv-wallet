package tokens

import (
	"context"
	"errors"
	"io"

	api "github.com/4chain-AG/gateway-overlay/pkg/open_api"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type TokenOverlayClient interface {
	VerifyAndSaveTokenTransfer(ctx context.Context, txHex *api.PutApiV1Bsv21TransferJSONRequestBody) error
}

type tokenOverlayClient struct {
	log zerolog.Logger
	api *api.Client
}

func NewTokenOverlayClient(logger *zerolog.Logger, overlayURL string, httpClient *resty.Client) (TokenOverlayClient, error) {
	api, err := api.NewClient(overlayURL, api.WithHTTPClient(httpClient.GetClient()))
	if err != nil {
		return nil, err
	}

	return &tokenOverlayClient{
		log: logger.With().Str("tokens", "token-overlay-client").Logger(),
		api: api,
	}, nil
}

func (c *tokenOverlayClient) VerifyAndSaveTokenTransfer(ctx context.Context, txHex *api.PutApiV1Bsv21TransferJSONRequestBody) error {
	resp, err := c.api.PutApiV1Bsv21Transfer(ctx, *txHex)
	if err != nil {
		c.log.Err(err).Ctx(ctx).Msg("Failed to send verify and save token transfer request")
		return err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 201:
		c.log.Info().Ctx(ctx).Msg("Overlay validate and register token transfer")
	case 204:
		c.log.Warn().Ctx(ctx).Msg("Overlay validate token transfer (already knew about it)")
	default:
		errorBody, _ := io.ReadAll(resp.Body)
		err = errors.New(string(errorBody))

		c.log.Err(err).Ctx(ctx).Msg("Failed register token transfer")
		return err
	}

	return nil
}
