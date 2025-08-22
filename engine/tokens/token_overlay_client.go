package tokens

import (
	"context"
	"errors"
	"io"
	"net/http"

	api "github.com/4chain-AG/gateway-overlay/pkg/open_api"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type APIVersion int

const (
	APIV1 APIVersion = iota
	APIV2
)

type TransferRequest struct {
	FeeVouts      *[]int `json:"fee_vouts,omitempty"`
	Hex           string `json:"hex"`
	ReceiverID    string `json:"receiver_id"`
	ReceiverVouts *[]int `json:"receiver_vouts,omitempty"`

	// SenderID envelope
	SenderID    string `json:"sender_id"`
	SenderVouts *[]int `json:"sender_vouts,omitempty"`

	AssetID string `json:"-"`
}

type TokenOverlayClient interface {
	VerifyAndSaveTokenTransfer(ctx context.Context, txHex *TransferRequest) error
}

type tokenOverlayClient struct {
	log        zerolog.Logger
	api        *api.Client
	apiVersion APIVersion
}

func NewTokenOverlayClient(logger *zerolog.Logger, overlayURL string, httpClient *resty.Client, apiVersion APIVersion) (TokenOverlayClient, error) {
	api, err := api.NewClient(overlayURL, api.WithHTTPClient(httpClient.GetClient()))
	if err != nil {
		return nil, err
	}

	return &tokenOverlayClient{
		log:        logger.With().Str("tokens", "token-overlay-client").Logger(),
		api:        api,
		apiVersion: apiVersion,
	}, nil
}

func (c *tokenOverlayClient) VerifyAndSaveTokenTransfer(ctx context.Context, transferReq *TransferRequest) error {
	var resp *http.Response
	var err error

	if c.apiVersion == APIV1 {
		resp, err = c.api.PutApiV1Bsv21Transfer(ctx, c.toV1TransferBody(transferReq))
	} else {
		resp, err = c.api.PutApiV2CoinAssetIdTransfer(ctx, transferReq.AssetID, c.toV2TransferBody(transferReq))
	}

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

func (c *tokenOverlayClient) toV1TransferBody(req *TransferRequest) api.PutApiV1Bsv21TransferJSONRequestBody {
	return api.PutApiV1Bsv21TransferJSONRequestBody{
		Hex:           req.Hex,
		FeeVouts:      req.FeeVouts,
		ReceiverId:    req.ReceiverID,
		ReceiverVouts: req.ReceiverVouts,
		SenderId:      req.SenderID,
		SenderVouts:   req.SenderVouts,
	}
}

func (c *tokenOverlayClient) toV2TransferBody(req *TransferRequest) api.PutApiV2CoinAssetIdTransferJSONRequestBody {
	return api.PutApiV2CoinAssetIdTransferJSONRequestBody{
		Hex:           req.Hex,
		FeeVouts:      req.FeeVouts,
		ReceiverId:    req.ReceiverID,
		ReceiverVouts: req.ReceiverVouts,
		SenderId:      req.SenderID,
		SenderVouts:   req.SenderVouts,
	}
}
