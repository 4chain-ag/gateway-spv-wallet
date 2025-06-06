package accesskeys

import (
	"net/http"

	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"github.com/bitcoin-sv/spv-wallet/mappings"
	"github.com/bitcoin-sv/spv-wallet/server/reqctx"
	"github.com/gin-gonic/gin"
)

// get will get an existing model
// Get access key godoc
// @Summary		Get access key
// @Description	Get access key
// @Tags		Access-key
// @Produce		json
// @Param		id path string true "id of the access key"
// @Success		200	{object} response.AccessKey "AccessKey with given id"
// @Failure		400	"Bad request - Missing required field: id"
// @Failure		403	"Forbidden - Access key is not owned by the user"
// @Failure 	500	"Internal server error - Error while getting access key"
// @Router		/api/v1/users/current/keys/{id} [get]
// @Security	x-auth-xpub
func get(c *gin.Context, userContext *reqctx.UserContext) {
	id := c.Params.ByName("id")
	reqXPubID := userContext.GetXPubID()

	logger := reqctx.Logger(c)

	if id == "" {
		spverrors.ErrorResponse(c, spverrors.ErrMissingFieldID, logger)
		return
	}

	// Get access key
	accessKey, err := reqctx.Engine(c).GetAccessKey(
		c.Request.Context(), reqXPubID, id,
	)
	if err != nil {
		spverrors.ErrorResponse(c, err, logger)
		return
	}

	if accessKey.XpubID != reqXPubID {
		spverrors.ErrorResponse(c, spverrors.ErrAuthorization, logger)
		return
	}

	contract := mappings.MapToAccessKeyContract(accessKey)
	c.JSON(http.StatusOK, contract)
}
