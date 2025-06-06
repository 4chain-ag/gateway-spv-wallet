package txmodels

// TxStatus represents possible statuses of stored transaction model
type TxStatus string

// List of transaction statuses
const (
	TxStatusCreated     TxStatus = "CREATED"
	TxStatusBroadcasted TxStatus = "BROADCASTED"
	TxStatusMined       TxStatus = "MINED"
	TxStatusReverted    TxStatus = "REVERTED"
	TxStatusProblematic TxStatus = "PROBLEMATIC"
)
