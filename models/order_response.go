package models

import "strings"

// OrderResponse distinguishes accepted orders from exchange rejections.
type OrderResponse interface{ orderResponse() }

func (*AcceptedOrder) orderResponse() {}
func (*RejectedOrder) orderResponse() {}

// NormalizeOrderResponse interprets the complete exchange response, not HTTP status alone.
func NormalizeOrderResponse(raw RawOrderResponse) OrderResponse {
	if raw.Success && raw.ErrorMsg == "" && raw.OrderID != "" && (raw.Status == "live" || raw.Status == "matched" || raw.Status == "delayed") {
		hashes := make([]TransactionHash, len(raw.TransactionsHashes))
		for i, value := range raw.TransactionsHashes {
			hashes[i] = TransactionHash(value)
		}
		return &AcceptedOrder{Ok: true, OrderID: OrderID(raw.OrderID), Status: raw.Status, MakingAmount: raw.MakingAmount, TakingAmount: raw.TakingAmount, TradeIDs: raw.TradeIDs, TransactionsHashes: hashes}
	}
	// Wire messages follow pinned upstream 7032ed06175b0bbea77b2d9e898bd9f252af2041.
	// Unknown messages are retained verbatim and classified as unknown.
	code := map[string]string{
		"the market is not yet ready to process new orders": "market_not_ready",
		"invalid nonce": "invalid_nonce", "invalid expiration": "invalid_expiration",
		"invalid post-only order: order crosses book":                                                              "post_only_would_cross",
		"post-only mode: only post-only orders and cancels are allowed":                                            "post_only_mode",
		"order couldn't be fully filled. FOK orders are fully filled or killed.":                                   "fok_not_filled",
		"no orders found to match with FAK order. FAK orders are partially filled or killed if no match is found.": "fak_not_filled",
	}[raw.ErrorMsg]
	if raw.Status == "unmatched" {
		code = "unmatched"
	}
	if code == "" && strings.Contains(raw.ErrorMsg, "not enough balance / allowance") {
		code = "not_enough_balance"
	}
	if code == "" {
		code = "unknown"
	}
	message := raw.ErrorMsg
	if message == "" {
		message = "Unknown order failure"
	}
	return &RejectedOrder{Code: code, Message: message}
}
