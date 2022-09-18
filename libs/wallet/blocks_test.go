package wallet

import (
	"strings"
	"testing"

	"github.com/appditto/pippin_nano_wallet/libs/utils"
	"github.com/appditto/pippin_nano_wallet/libs/utils/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestGetBlockFromDatabase(t *testing.T) {
	// Predictable seed
	seed, _ := utils.GenerateSeed(strings.NewReader("42f567f406715877a8d60c6890b4655e8e0fe64f3f82089fe76899f180cd4021"))

	wallet, err := MockWallet.WalletCreate(seed)
	assert.Nil(t, err)

	assert.Equal(t, false, wallet.Encrypted)
	assert.Equal(t, seed, wallet.Seed)

	// Create an account and an adhoc account
	account, err := MockWallet.AccountCreate(wallet, nil)
	assert.Nil(t, err)

	_, priv, _ := ed25519.GenerateKey(strings.NewReader("1f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040d"))
	adhocAcct, err := MockWallet.AdhocAccountCreate(wallet, priv)
	assert.Nil(t, err)

	// Create some blocks for each
	// Create a block object
	_, err = MockWallet.DB.Block.Create().SetAccount(account).SetBlock(map[string]interface{}{
		"block": "hello",
	}).SetBlockHash("abc").SetSubtype("send").SetSendID("abc").Save(MockWallet.Ctx)
	assert.Nil(t, err)
	_, err = MockWallet.DB.Block.Create().SetAccount(adhocAcct).SetBlock(map[string]interface{}{
		"block": "world",
	}).SetBlockHash("def").SetSubtype("change").SetSendID("def").Save(MockWallet.Ctx)
	assert.Nil(t, err)

	// Retrieve block
	block, err := MockWallet.GetBlockFromDatabase(wallet, account.Address, "abc")
	assert.Nil(t, err)
	assert.Equal(t, "hello", block.Block["block"])
	block, err = MockWallet.GetBlockFromDatabase(wallet, adhocAcct.Address, "def")
	assert.Nil(t, err)
	assert.Equal(t, "world", block.Block["block"])
	block, err = MockWallet.GetBlockFromDatabase(wallet, account.Address, "nonexistent")
	assert.ErrorIs(t, err, ErrBlockNotFound)
}

// func TestReceiveBlockCreate(t *testing.T) {
// 	httpmock.Activate()
// 	defer httpmock.DeactivateAndReset()

// 	httpmock.RegisterResponder("POST", "/mockrpcendpoint",
// 		func(req *http.Request) (*http.Response, error) {
// 			var pr requests.BaseRequest
// 			json.NewDecoder(req.Body).Decode(&pr)
// 			if pr.Action == "block_info" {
// 				var js map[string]interface{}
// 				json.Unmarshal([]byte(mocks.BlockInfoResponseStr), &js)
// 				resp, err := httpmock.NewJsonResponse(200, js)
// 				return resp, err
// 			}
// 			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
// 				"error": "error",
// 			})
// 			return resp, err
// 		},
// 	)

// 	// Input inputs
// 	_, err := MockWallet.createReceiveBlock(nil, nil, "", nil)
// 	assert.ErrorIs(t, err, ErrInvalidWallet)
// 	_, err = MockWallet.createReceiveBlock(&ent.Wallet{}, nil, "", nil)
// 	assert.ErrorIs(t, err, ErrInvalidAccount)

// 	// Create a wallet
// 	seed, err := utils.GenerateSeed(strings.NewReader("9c0784e354217e282e8b0177db3bf18d74768d8adb1de88412c6c4d9fd33f407"))
// 	assert.Nil(t, err)
// 	wallet, err := MockWallet.WalletCreate(seed)
// 	assert.Nil(t, err)

// 	// Create an account
// 	account, err := MockWallet.AccountCreate(wallet, nil)
// 	assert.Nil(t, err)

// }
