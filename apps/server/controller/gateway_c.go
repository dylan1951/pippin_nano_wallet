package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/appditto/pippin_nano_wallet/libs/log"
	"golang.org/x/exp/slices"
)

var UNSUPPORTED_WALLET_ACTIONS = []string{"account_move", "account_remove", "receive_minimum", "receive_minimum_set", "search_pending", "search_pending_all", "wallet_add_watch", "wallet_export", "wallet_history", "wallet_ledger", "wallet_republish", "wallet_work_get", "work_get", "work_set"}

// This is called the "Gateway" because it's the entry point for all requests
// This API is intended to replace the nano node wallet RPCs
// https://docs.nano.org/commands/rpc-protocol/#wallet-rpcs
// It will:
// 1) Determine if the request is a supported wallet RPC, if so process it
// 2) If it's a wallet RPC we don't support, return error
// 3) Other requests with a correct signature go straight to the node
// The error messages and behavior are also intended to replace what the nano node returns
// The node isn't exactly great at returning errors, and the error messages are not very helpful
// But as we want to be a drop-in replacement we mimic the behavior
func (hc *HttpController) Gateway(w http.ResponseWriter, r *http.Request) {
	var baseRequest map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&baseRequest); err != nil {
		log.Errorf("Error unmarshalling http base request %s", err)
		ErrUnableToParseJson(w, r)
		return
	}

	if _, ok := baseRequest["action"]; !ok {
		ErrUnableToParseJson(w, r)
		return
	}

	action := strings.ToLower(fmt.Sprintf("%v", baseRequest["action"]))

	if slices.Contains(UNSUPPORTED_WALLET_ACTIONS, action) {
		ErrBadRequest(w, r, "not_implemented")
		return
	}

	switch action {
	case "wallet_create":
		hc.HandleWalletCreate(&baseRequest, w, r)
		return
	case "account_create":
		hc.HandleAccountCreate(&baseRequest, w, r)
		return
	case "accounts_create":
		hc.HandleAccountsCreate(&baseRequest, w, r)
		return
	case "account_list":
		hc.HandleAccountList(&baseRequest, w, r)
		return
	case "password_change":
		hc.HandlePasswordChange(&baseRequest, w, r)
		return
	case "password_enter":
		hc.HandlePasswordEnter(&baseRequest, w, r)
		return
	case "wallet_add":
		hc.HandleWalletAdd(&baseRequest, w, r)
		return
	case "wallet_locked":
		hc.HandleWalletLocked(&baseRequest, w, r)
		return
	case "wallet_lock":
		hc.HandleWalletLock(&baseRequest, w, r)
		return
	case "wallet_destroy":
		hc.HandleWalletDestroy(&baseRequest, w, r)
		return
	case "wallet_balances":
		hc.HandleWalletBalances(&baseRequest, w, r)
		return
	case "wallet_frontiers":
		hc.HandleWalletFrontiers(&baseRequest, w, r)
		return
	case "wallet_pending":
		hc.HandleWalletPending(&baseRequest, w, r)
		return
	case "work_generate":
		hc.HandleWorkGenerate(&baseRequest, w, r)
		return
	case "wallet_info":
		hc.HandleWalletInfo(&baseRequest, w, r)
		return
	case "wallet_contains":
		hc.HandleWalletContains(&baseRequest, w, r)
		return
	case "receive":
		hc.HandleReceiveRequest(&baseRequest, w, r)
		return
	case "receive_all":
		hc.HandleReceiveAllRequest(&baseRequest, w, r)
		return
	case "account_receive_all":
		hc.HandleAccountReceiveAllRequest(&baseRequest, w, r)
		return
	case "send":
		hc.HandleSendRequest(&baseRequest, w, r)
		return
	case "account_representative_set":
		hc.HandleAccountRepresentativeSetRequest(&baseRequest, w, r)
		return
	case "wallet_representative_set":
		hc.HandleWalletRepresentativeSetRequest(&baseRequest, w, r)
		return
	case "wallet_representative":
		hc.HandleWalletRepresentativeRequest(&baseRequest, w, r)
		return
	case "wallet_change_seed":
		hc.HandleWalletChangeSeedRequest(&baseRequest, w, r)
		return
	default:
		resp, err := hc.RpcClient.MakeRequest(baseRequest)
		if err != nil {
			ErrInternalServerError(w, r, "Error forwarding request to node")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}
