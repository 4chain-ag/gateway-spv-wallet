package stablecoins

import (
	"net/http"

	"github.com/bitcoin-sv/spv-wallet/engine"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"github.com/bitcoin-sv/spv-wallet/server/reqctx"
	"github.com/gin-gonic/gin"
)

// stablecoinTransferIntent validate transfer intent
// Paymail validate transfer intent
// @Summary		Validate transfer intent
// @Description	Validate transfer intent
// @Tags		Stablecoin Transfer
// @Produce		json
// @Param		StablecoinTransferIntent body Intent true "Transfer intent use to create outputs and validate transfer"
// @Success		200 {object} ValidationResponse "Transfer intent validation response"
// @Failure		400	"Bad request - Error while parsing SearchPaymails from request body"
// @Failure 	500	"Internal server error - Error while searching for paymail addresses"
// @Router		/bsvalias/transfer-intent [post]
func stablecoinTransferIntent(c *gin.Context) {
	logger := reqctx.Logger(c)
	engineInstance := reqctx.Engine(c)

	var requestBody *engine.Intent
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		spverrors.ErrorResponse(c, err, logger)
		return
	}

	resp, err := engineInstance.StablecoinTransferService().ValidateIntent(c.Request.Context(), engineInstance, requestBody)
	if err != nil {
		spverrors.ErrorResponse(c, err, logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}
