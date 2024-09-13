package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/appditto/pippin_nano_wallet/apps/server/models/requests"
	"github.com/appditto/pippin_nano_wallet/apps/server/models/responses"
	"github.com/appditto/pippin_nano_wallet/libs/log"
	"github.com/appditto/pippin_nano_wallet/libs/utils"
	"github.com/appditto/pippin_nano_wallet/libs/wallet"
	"github.com/go-chi/render"
	"github.com/mitchellh/mapstructure"
)

// Handle receive individual block
func (hc *HttpController) HandleReceiveRequest(rawRequest *map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	var receiveRequest requests.ReceiveRequest
	if err := mapstructure.Decode(rawRequest, &receiveRequest); err != nil {
		log.Errorf("Error unmarshalling receive request %s", err)
		ErrUnableToParseJson(w, r)
		return
	} else if receiveRequest.Wallet == "" || receiveRequest.Action == "" || receiveRequest.Block == "" {
		ErrUnableToParseJson(w, r)
		return
	}

	// See if wallet exists
	dbWallet := hc.WalletExists(receiveRequest.Wallet, w, r)
	if dbWallet == nil {
		return
	}

	// Validate account
	_, err := utils.AddressToPub(receiveRequest.Account, hc.Wallet.Config.Wallet.Banano)
	if err != nil {
		ErrInvalidAccount(w, r)
		return
	}

	// Accounts list
	resp, err := hc.Wallet.CreateAndPublishReceiveBlock(dbWallet, receiveRequest.Account, receiveRequest.Block, receiveRequest.Work, receiveRequest.BpowKey)
	if err != nil {
		ErrBadRequest(w, r, err.Error())
		return
	}

	blockResponse := responses.BlockResponse{
		Block: resp,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &blockResponse)
}

// Handle receive all blocks for a specific account
func (hc *HttpController) HandleAccountReceiveAllRequest(rawRequest *map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	// mapstructure decode
	var accountReceiveAllRequest requests.AccountReceiveAllRequest
	if err := mapstructure.Decode(rawRequest, &accountReceiveAllRequest); err != nil {
		log.Errorf("Error unmarshalling account_receive_all request %s", err)
		ErrUnableToParseJson(w, r)
		return
	}

	// See if wallet exists
	dbWallet := hc.WalletExists(accountReceiveAllRequest.Wallet, w, r)
	if dbWallet == nil {
		return
	}

	// Validate account
	_, err := utils.AddressToPub(accountReceiveAllRequest.Account, hc.Wallet.Config.Wallet.Banano)
	if err != nil {
		ErrBadRequest(w, r, fmt.Sprintf("Invalid account %s", accountReceiveAllRequest.Account))
		return
	}

	receivedCount := 0
	resp, err := hc.Wallet.ReceiveAllBlocks(dbWallet, accountReceiveAllRequest.Account, accountReceiveAllRequest.BpowKey)
	if err != nil {
		ErrInternalServerError(w, r, err.Error())
		return
	}
	receivedCount += resp

	response := responses.ReceiveAllResponse{
		Received: receivedCount,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &response)
}

// Handle receive all blocks in entire wallet
func (hc *HttpController) HandleReceiveAllRequest(rawRequest *map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	request := hc.DecodeBaseRequest(rawRequest, w, r)
	if request == nil {
		return
	}

	// See if wallet exists
	dbWallet := hc.WalletExists(request.Wallet, w, r)
	if dbWallet == nil {
		return
	}

	// Accounts list
	_, accounts, err := hc.Wallet.AccountsList(dbWallet, 0)
	if errors.Is(err, wallet.ErrWalletLocked) {
		ErrWalletLocked(w, r)
		return
	} else if err != nil {
		ErrInternalServerError(w, r, err.Error())
		return
	}

	receivedCount := 0
	for _, account := range accounts {
		resp, err := hc.Wallet.ReceiveAllBlocks(dbWallet, account, request.BpowKey)
		if err != nil {
			ErrInternalServerError(w, r, err.Error())
			return
		}
		receivedCount += resp
	}

	resp := responses.ReceiveAllResponse{
		Received: receivedCount,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &resp)
}

// Handle send block
func (hc *HttpController) HandleSendRequest(rawRequest *map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	var sendRequest requests.SendRequest
	if err := mapstructure.Decode(rawRequest, &sendRequest); err != nil {
		log.Errorf("Error unmarshalling send request %s", err)
		ErrUnableToParseJson(w, r)
		return
	} else if sendRequest.Wallet == "" || sendRequest.Action == "" || sendRequest.Amount == "" || sendRequest.Destination == "" {
		ErrUnableToParseJson(w, r)
		return
	}

	// See if wallet exists
	dbWallet := hc.WalletExists(sendRequest.Wallet, w, r)
	if dbWallet == nil {
		return
	}

	// Validate accounts
	_, err := utils.AddressToPub(sendRequest.Source, hc.Wallet.Config.Wallet.Banano)
	if err != nil {
		ErrBadRequest(w, r, fmt.Sprintf("Invalid source account %s", sendRequest.Source))
		return
	}
	_, err = utils.AddressToPub(sendRequest.Destination, hc.Wallet.Config.Wallet.Banano)
	if err != nil {
		ErrBadRequest(w, r, fmt.Sprintf("Invalid destination account %s", sendRequest.Destination))
		return
	}

	fmt.Printf("calling CreateAndPublishSendBlock\n")
	// Do the send
	resp, err := hc.Wallet.CreateAndPublishSendBlock(dbWallet, sendRequest.Amount, sendRequest.Source, sendRequest.Destination, sendRequest.ID, sendRequest.Work, sendRequest.BpowKey)
	if err != nil {
		fmt.Printf("encountered error in CreateAndPublishSendBlock\n")
		ErrBadRequest(w, r, err.Error())
		return
	}

	blockResponse := responses.BlockResponse{
		Block: resp,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &blockResponse)
}

// Handle rep change
func (hc *HttpController) HandleAccountRepresentativeSetRequest(rawRequest *map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	var changeRequest requests.AccountRepresentativeSetRequest
	if err := mapstructure.Decode(rawRequest, &changeRequest); err != nil {
		log.Errorf("Error unmarshalling representative set request %s", err)
		ErrUnableToParseJson(w, r)
		return
	} else if changeRequest.Wallet == "" || changeRequest.Action == "" || changeRequest.Representative == "" {
		ErrUnableToParseJson(w, r)
		return
	}

	// See if wallet exists
	dbWallet := hc.WalletExists(changeRequest.Wallet, w, r)
	if dbWallet == nil {
		return
	}

	// Validate accounts
	_, err := utils.AddressToPub(changeRequest.Account, hc.Wallet.Config.Wallet.Banano)
	if err != nil {
		ErrBadRequest(w, r, "Invalid account")
		return
	}
	_, err = utils.AddressToPub(changeRequest.Representative, hc.Wallet.Config.Wallet.Banano)
	if err != nil {
		ErrBadRequest(w, r, "Invalid representative account")
		return
	}

	// Do the send
	resp, err := hc.Wallet.CreateAndPublishChangeBlock(dbWallet, changeRequest.Account, changeRequest.Representative, changeRequest.Work, changeRequest.BpowKey, false)
	if err != nil {
		ErrBadRequest(w, r, err.Error())
		return
	}

	blockResponse := responses.BlockResponse{
		Block: resp,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &blockResponse)
}
