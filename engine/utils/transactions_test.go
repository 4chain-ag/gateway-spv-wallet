package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfEf(t *testing.T) {
	t.Run("valid tx", func(t *testing.T) {
		txHex := "010000000000000000ef02d6f6570aa3e4a068f3548474a592246c942a37c0a899446ac863c79bfc152369000000006b483045022100d1348eda6066dd75b0adb436ab874f1b9982515f108c7b6c1588652353647f53022012f9f60bf11edff914d0fed6a22eedaa48824e0733f88ab66c0cbe907a427eb9412102a1ad181d0d0d68d7d4d42907f9f718cb7769eec5093829ad39642c2bdd09619dffffffff0100000000000000aa76a91440adec2fa5295b9fcfeb1ffc5f1cc666cabc0e1888ac0063036f726451126170706c69636174696f6e2f6273762d3230004c737b22616d74223a223130222c226964223a22633963656332356132303361623238646635613262303536633436613462633439363735613434653962643139323232383031316139396665353137386238375f30222c226f70223a227472616e73666572222c2270223a226273762d3230227d68d6f6570aa3e4a068f3548474a592246c942a37c0a899446ac863c79bfc152369010000006a473044022066720bc2093fd7e6449f7a1fa6d5708d4f5aab50253e4e98c7f1634c5cce16df02204b4465eb071c986b355426e79c7dd4e4060c077b17a10bf1fabd6c6870638a9f41210379c5b43e0b96bd115a01e2db01c5caedb77819ac8f3da645f0d210ae63c8d560ffffffff03000000000000001976a914c33c4b35f5adbb562384ea10724986ebf7e2d2fe88ac020100000000000000aa76a91431d7f641091e23e5f8148b198b8908b4c4494f9088ac0063036f726451126170706c69636174696f6e2f6273762d3230004c737b22616d74223a223130222c226964223a22633963656332356132303361623238646635613262303536633436613462633439363735613434653962643139323232383031316139396665353137386238375f30222c226f70223a227472616e73666572222c2270223a226273762d3230227d6802000000000000001976a914702658424302c364f023c0e31946bbab1d4a415688ac00000000"
		assert.Equal(t, true, IsEf(txHex))
	})

	t.Run("invalid tx", func(t *testing.T) {
		txHex := "01000000016bc654c41aaed214ad2c9b85354749f24ca9989eb9d362c0ee4c6dcfd89ec20e000000006b483045022100ff23565511fad5d1fba18f9d96e93def0c779bfb04859a8eb5f0b2155bda4942022037c731e2b8f5a9e01d487d96b5edf92233c36cb01ea9a54aaf57df3d1fcedbbd412102e31115b0e5acb4721b4f53b902e6fc55fb6b17ef513fefba00aa8a0a3a57dbf0ffffffff0201000000000000007d76a914139ef757989cb4dfcb6c333d815eb728ef9d09eb88ac0063036f726451126170706c69636174696f6e2f6273762d323000477b22616d74223a223130222c226f70223a226465706c6f792b6d696e74222c2270223a226273762d3230222c2273796d223a22676174657761792d746f6b656e2d74657374227d6804000000000000001976a91425dc22d589aaa29addbed75a755d70b047665b7e88ac00000000"
		assert.Equal(t, false, IsEf(txHex))
	})
}
