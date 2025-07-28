package transfer

import (
	routes "github.com/bitcoin-sv/spv-wallet/server/handlers"
)

// RegisterRoutes creates the specific package routes in RESTful style
func RegisterRoutes(handlersManager *routes.Manager) {
	group := handlersManager.Group(routes.GroupRoot, "/bsvalias")
	group.POST("/transfer-intent", transferIntent)
	group.POST("/transfer", transfer)
}
