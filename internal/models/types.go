package models

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type SwapRequest struct {
	RequestID    string         `json:"requestId"`
	FromChainID  int64          `json:"fromChainId"`
	ToChainID    int64          `json:"toChainId"`
	TokenAddress common.Address `json:"tokenAddress"`
	Amount       *big.Int       `json:"amount"`
	Recipient    common.Address `json:"recipient"`
	Timestamp    time.Time      `json:"timestamp"`
}

type SwapStatus struct {
	RequestID    string    `json:"requestId"`
	Status       string    `json:"status"`
	FromChainID  int64     `json:"fromChainId"`
	ToChainID    int64     `json:"toChainId"`
	SourceTxHash string    `json:"sourceTxHash,omitempty"`
	TargetTxHash string    `json:"targetTxHash,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type QueueStatus struct {
	Length        int `json:"length"`
	MaxSize       int `json:"maxSize"`
	ActiveBatches int `json:"activeBatches"`
}
