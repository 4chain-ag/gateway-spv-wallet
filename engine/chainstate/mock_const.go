package chainstate

import "github.com/bitcoin-sv/spv-wallet/engine/utils"

const (
	// Dummy transaction data
	broadcastExample1TxID        = "15d31d00ed7533a83d7ab206115d7642812ec04a2cbae4248365febb82576ff3"
	broadcastExample1TxHex       = "0100000001018d7ab1a0f0253120a0cb284e4170b47e5f83f70faaba5b0b55bbeeef624b45010000006b483045022100d5b0dddf76da9088e21cf1277f064dc7832c3da666732f003ee48f2458142e9a02201fe725a1c455b2bd964779391ae105b87730881f211cd299ca36d70d74d715ab412103673dffd80561b87825658f74076da805c238e8c47f25b5d804893c335514d074ffffffff02c4090000000000001976a914777242b335bc7781f43e1b05c60d8c2f2d08b44c88ac962e0000000000001976a91467d93a70ac575e15abb31bc8272a00ab1495d48388ac00000000"
	notFoundExample1TxID         = "918c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b"
	onChainExample1BlockHash     = "0000000000000000015122781ab51d57b26a09518630b882f67f1b08d841979d"
	onChainExample1BlockHeight   = int64(723229)
	onChainExample1Confirmations = int64(314)
	onChainExample1TxHex         = "01000000025b7439a0c9effa3f19d0e441d2eea596e44a8c49240b6e389c29498285f92ad3010000006a4730440220482c1c896678d7307e1de35cef2aae4907f2684617a26d8abd24c444d527c80d02204c550f8f9d69b9cf65780e2e066041750261702639d02605a2eb694ade4ca1d64121029ce7958b2aa3c627334f50bb810c678e2b284db0ef6f7d067f7fccfa05d0f095ffffffff1998b0e4955e1d8ba976d943c43f32e143ba90e805f0e882d3b8edc0f7473b77020000006a47304402204beb486e5d99a15d4d2267e328abb5466a05fdc20d64903d0ace1c4fabb71a34022024803ae9e18b3c11683b2ff2b5fb4ca973a22fdd390f6ab1f99396604a3f06af4121038ea0f258fb838b5193e9739ddd808bb97aaab52a60ba8a83958b13109ab183ccffffffff030000000000000000fd8901006a0372756e0105004d7d017b22696e223a312c22726566223a5b22653864393134303764643461646164363366333739353032303861383532653562306334383037333563656235346133653334333539346163313839616331625f6f31222c22376135346462326162303030306161303035316134383230343162336135653761636239386333363135363863623334393063666564623066653161356438385f6f33225d2c226f7574223a5b2233356463303036313539393333623438353433343565663663633363366261663165666462353263343837313933386632366539313034343632313562343036225d2c2264656c223a5b5d2c22637265223a5b5d2c2265786563223a5b7b226f70223a2243414c4c222c2264617461223a5b7b22246a6967223a307d2c22757064617465222c5b7b22246a6967223a317d2c7b2267726164756174696f6e506f736974696f6e223a6e756c6c2c226c6576656c223a382c226e616d65223a22e38395e383abe38380222c227870223a373030307d5d5d7d5d7d11010000000000001976a914058cae340a2ef8fd2b43a074b75fb6b38cb2765788acd4020000000000001976a914160381a3811b474ff77f31f64f4e57a5bb5ebf1788ac00000000"
	onChainExample1TxID          = "908c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b"
	onChainExampleArcTxID        = "a11b9e1ee08e264f9add02e4afa40dad3c00a23f250ac04449face095c68fab7"

	// API key
	// testDummyKey = "test-dummy-api-key-value" //nolint:gosec // this is a dummy key

	// Signatures
	matterCloudSig1 = "30450221008003e78e2154e9686a2bb864a811e7ec950093f273f94d222fa87c81a5daf8f6022018e1098ad2a3f1adc431ddd375c06c867d79e484710dc87ac6847b3a2f4909d2"
	gorillaPoolSig1 = "3045022100bfa217db8eb1a520db05877c724ce4150d16ff0ff93165e1d2b6498a04a3da3102203cb00b39de31c3ed12456c212cc2e3860928db875eee5f843bce7b9f300beef0"

	// Defaults for request payloads
	utf8Type            = "UTF-8"
	applicationJSONType = "application/json"
)

// MockDefaultFee is a mock default fee used for assertions
var MockDefaultFee = &utils.FeeUnit{
	Satoshis: 1,
	Bytes:    20,
}
