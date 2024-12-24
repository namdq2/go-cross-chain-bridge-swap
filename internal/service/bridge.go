package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/namdq2/go-cross-chain-bridge-swap/internal/models"
	"github.com/namdq2/go-cross-chain-bridge-swap/internal/processor"
)

type Config struct {
	Chain1RPC   string
	Chain2RPC   string
	Chain1ID    int64
	Chain2ID    int64
	BridgeAddr1 string
	BridgeAddr2 string
}

type BridgeService struct {
	config         Config
	batchProcessor *processor.BatchProcessor
	walletPool     *processor.WalletPool
	db             *models.Database
}

func NewBridgeService(config Config, privateKeys []string, db *models.Database) (*BridgeService, error) {
	walletPool, err := processor.NewWalletPool(privateKeys)
	if err != nil {
		return nil, err
	}

	service := &BridgeService{
		config:     config,
		walletPool: walletPool,
		db:         db,
	}

	service.batchProcessor = processor.NewBatchProcessor(walletPool, db)
	return service, nil
}

func (s *BridgeService) InitiateSwap(ctx context.Context, req *models.SwapRequest) (*models.SwapStatus, error) {
	// Validate request
	if err := s.validateSwapRequest(req); err != nil {
		return nil, err
	}

	// Save to database
	swap, err := s.db.CreateSwap(ctx, &models.SwapRequest{
		RequestID:    req.RequestID,
		FromChainID:  req.FromChainID,
		ToChainID:    req.ToChainID,
		TokenAddress: req.TokenAddress,
		Amount:       req.Amount,
		Recipient:    req.Recipient,
		Timestamp:    time.Now(),
	})
	if err != nil {
		return nil, err
	}

	// Add to batch processor
	s.batchProcessor.AddRequest(req)

	return &models.SwapStatus{
		RequestID:   req.RequestID,
		Status:      "pending",
		FromChainID: req.FromChainID,
		ToChainID:   req.ToChainID,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *BridgeService) validateSwapRequest(req *models.SwapRequest) error {
	// Validate chain IDs
	if req.FromChainID != s.config.Chain1ID && req.FromChainID != s.config.Chain2ID {
		return fmt.Errorf("invalid source chain ID")
	}
	if req.ToChainID != s.config.Chain1ID && req.ToChainID != s.config.Chain2ID {
		return fmt.Errorf("invalid destination chain ID")
	}
	if req.FromChainID == req.ToChainID {
		return fmt.Errorf("source and destination chains must be different")
	}

	// Validate token address
	if !common.IsHexAddress(req.TokenAddress.Hex()) {
		return fmt.Errorf("invalid token address")
	}

	// Validate amount
	if req.Amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Validate recipient
	if !common.IsHexAddress(req.Recipient.Hex()) {
		return fmt.Errorf("invalid recipient address")
	}

	return nil
}

func (s *BridgeService) GetSwapStatus(ctx context.Context, requestID string) (*models.SwapStatus, error) {
	return s.db.GetSwapStatus(ctx, requestID)
}

func (s *BridgeService) GetQueueStatus(ctx context.Context) *models.QueueStatus {
	return &models.QueueStatus{
		Length:        s.batchProcessor.GetCurrentBatchSize(),
		MaxSize:       processor.BATCH_SIZE,
		ActiveBatches: s.batchProcessor.GetActiveBatchCount(),
	}
}
