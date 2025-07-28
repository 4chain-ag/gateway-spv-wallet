package transfer

import (
	"net/http"

	"github.com/bitcoin-sv/spv-wallet/engine"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"github.com/bitcoin-sv/spv-wallet/server/reqctx"
	"github.com/gin-gonic/gin"
)

// transfer incoming transfer
// Paymail incoming transfer
// @Summary		Incoming transfer
// @Description	Incoming transfer
// @Tags		Transfer
// @Produce		json
// @Param		TransferData body Transfer true "Transfer info"
// @Success		200 {object} ValidationResponse "Transfer intent validation response"
// @Failure		400	"Bad request - Error while parsing SearchPaymails from request body"
// @Failure 	500	"Internal server error - Error while searching for paymail addresses"
// @Router		/bsvalias/transfer [post]
func transfer(c *gin.Context) {
	logger := reqctx.Logger(c)
	engineInstance := reqctx.Engine(c)

	var requestBody *engine.Transfer
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		spverrors.ErrorResponse(c, err, logger)
		return
	}

	tx, err := engineInstance.TransferService().IncomingTransfer(c.Request.Context(), engineInstance, *requestBody)
	if err != nil {
		spverrors.ErrorResponse(c, err, logger)
		return
	}

	c.JSON(http.StatusOK, tx)
}
