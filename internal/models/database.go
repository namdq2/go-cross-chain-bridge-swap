// models/database.go

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
)

var (
	ErrSwapNotFound  = errors.New("swap not found")
	ErrBatchNotFound = errors.New("batch not found")
	ErrInvalidStatus = errors.New("invalid status")
)

type Database struct {
	db *sql.DB
}

type SwapRequest struct {
	RequestID    string
	FromChainID  int64
	ToChainID    int64
	TokenAddress common.Address
	Amount       string
	Recipient    common.Address
	Status       string
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Batch struct {
	ID            int64
	BatchID       string
	WalletAddress string
	ChainID       int64
	SourceTxHash  *string
	TargetTxHash  *string
	Status        string
	GasPrice      *string
	GasUsed       *int64
	BlockNumber   *int64
	ErrorMessage  *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type HotWallet struct {
	ID                    int64
	Address               string
	ChainID               int64
	Nonce                 int64
	LastUsedAt            *time.Time
	IsActive              bool
	TotalProcessedBatches int
	TotalProcessedVolume  string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ChainConfig struct {
	ID                    int64
	ChainID               int64
	ChainType             string
	RPCUrl                string
	BridgeAddress         string
	RequiredConfirmations int
	MaxGasPrice           *string
	IsActive              bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	return &Database{db: db}, nil
}

// Swap related functions
func (db *Database) CreateSwap(ctx context.Context, swap *SwapRequest) error {
	query := `
        INSERT INTO swaps (
            request_id, from_chain_id, to_chain_id, 
            token_address, amount, recipient, 
            status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at
    `

	return db.db.QueryRowContext(
		ctx,
		query,
		swap.RequestID,
		swap.FromChainID,
		swap.ToChainID,
		swap.TokenAddress.Hex(),
		swap.Amount,
		swap.Recipient.Hex(),
		swap.Status,
	).Scan(&swap.CreatedAt, &swap.UpdatedAt)
}

func (db *Database) GetSwapByRequestID(ctx context.Context, requestID string) (*SwapRequest, error) {
	query := `
        SELECT 
            request_id, from_chain_id, to_chain_id,
            token_address, amount, recipient,
            status, error_message, created_at, updated_at
        FROM swaps 
        WHERE request_id = $1
    `

	swap := &SwapRequest{}
	err := db.db.QueryRowContext(ctx, query, requestID).Scan(
		&swap.RequestID,
		&swap.FromChainID,
		&swap.ToChainID,
		&swap.TokenAddress,
		&swap.Amount,
		&swap.Recipient,
		&swap.Status,
		&swap.ErrorMessage,
		&swap.CreatedAt,
		&swap.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSwapNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting swap: %v", err)
	}

	return swap, nil
}

func (db *Database) UpdateSwapStatus(ctx context.Context, requestID string, status string, errorMsg *string) error {
	query := `
        UPDATE swaps 
        SET status = $1, error_message = $2, updated_at = NOW()
        WHERE request_id = $3
    `

	result, err := db.db.ExecContext(ctx, query, status, errorMsg, requestID)
	if err != nil {
		return fmt.Errorf("error updating swap status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return ErrSwapNotFound
	}

	return nil
}

// Batch related functions
func (db *Database) CreateBatch(ctx context.Context, batch *Batch) error {
	query := `
        INSERT INTO batches (
            batch_id, wallet_address, chain_id,
            status
        ) VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `

	return db.db.QueryRowContext(
		ctx,
		query,
		batch.BatchID,
		batch.WalletAddress,
		batch.ChainID,
		batch.Status,
	).Scan(&batch.ID, &batch.CreatedAt, &batch.UpdatedAt)
}

func (db *Database) AddSwapsToBatch(ctx context.Context, batchID int64, swapIDs []string) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert batch_swaps mappings
	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO batch_swaps (batch_id, swap_id)
        VALUES ($1, $2)
    `)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	for _, swapID := range swapIDs {
		_, err = stmt.ExecContext(ctx, batchID, swapID)
		if err != nil {
			return fmt.Errorf("error inserting batch_swap: %v", err)
		}
	}

	// Update swaps status
	_, err = tx.ExecContext(ctx, `
        UPDATE swaps 
        SET status = 'queued', updated_at = NOW()
        WHERE request_id = ANY($1)
    `, pq.Array(swapIDs))
	if err != nil {
		return fmt.Errorf("error updating swaps status: %v", err)
	}

	return tx.Commit()
}

func (db *Database) UpdateBatchStatus(ctx context.Context, batchID string, status string, txHash *string, gasUsed *int64, gasPrice *string, errorMsg *string) error {
	query := `
        UPDATE batches 
        SET status = $1, 
            source_tx_hash = $2,
            gas_used = $3,
            gas_price = $4,
            error_message = $5,
            updated_at = NOW()
        WHERE batch_id = $6
    `

	result, err := db.db.ExecContext(
		ctx, query,
		status, txHash, gasUsed, gasPrice, errorMsg, batchID,
	)
	if err != nil {
		return fmt.Errorf("error updating batch status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return ErrBatchNotFound
	}

	return nil
}

// Hot wallet related functions
func (db *Database) GetAvailableWallet(ctx context.Context, chainID int64) (*HotWallet, error) {
	query := `
        UPDATE hot_wallets
        SET last_used_at = NOW()
        WHERE id = (
            SELECT id
            FROM hot_wallets
            WHERE chain_id = $1
            AND is_active = true
            ORDER BY last_used_at NULLS FIRST, total_processed_batches
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        )
        RETURNING id, address, chain_id, nonce, last_used_at,
                  is_active, total_processed_batches, total_processed_volume,
                  created_at, updated_at
    `

	wallet := &HotWallet{}
	err := db.db.QueryRowContext(ctx, query, chainID).Scan(
		&wallet.ID,
		&wallet.Address,
		&wallet.ChainID,
		&wallet.Nonce,
		&wallet.LastUsedAt,
		&wallet.IsActive,
		&wallet.TotalProcessedBatches,
		&wallet.TotalProcessedVolume,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("no available wallets")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting wallet: %v", err)
	}

	return wallet, nil
}

func (db *Database) UpdateWalletNonce(ctx context.Context, walletID int64, nonce int64) error {
	query := `
        UPDATE hot_wallets
        SET nonce = $1, updated_at = NOW()
        WHERE id = $2
    `

	result, err := db.db.ExecContext(ctx, query, nonce, walletID)
	if err != nil {
		return fmt.Errorf("error updating wallet nonce: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("wallet not found")
	}

	return nil
}

// Chain config related functions
func (db *Database) GetChainConfig(ctx context.Context, chainID int64) (*ChainConfig, error) {
	query := `
        SELECT 
            id, chain_id, chain_type, rpc_url,
            bridge_address, required_confirmations,
            max_gas_price, is_active, created_at, updated_at
        FROM chain_configs
        WHERE chain_id = $1 AND is_active = true
    `

	config := &ChainConfig{}
	err := db.db.QueryRowContext(ctx, query, chainID).Scan(
		&config.ID,
		&config.ChainID,
		&config.ChainType,
		&config.RPCUrl,
		&config.BridgeAddress,
		&config.RequiredConfirmations,
		&config.MaxGasPrice,
		&config.IsActive,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("chain config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting chain config: %v", err)
	}

	return config, nil
}

// Statistics and monitoring
func (db *Database) GetSwapStatistics(ctx context.Context, fromTime time.Time) (map[string]interface{}, error) {
	query := `
        SELECT 
            COUNT(*) as total_swaps,
            COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_swaps,
            COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_swaps,
            AVG(EXTRACT(EPOCH FROM (updated_at - created_at))) as avg_processing_time
        FROM swaps
        WHERE created_at >= $1
    `

	var stats = make(map[string]interface{})
	var avgProcessingTime *float64

	err := db.db.QueryRowContext(ctx, query, fromTime).Scan(
		&stats["total_swaps"],
		&stats["completed_swaps"],
		&stats["failed_swaps"],
		&avgProcessingTime,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting swap statistics: %v", err)
	}

	if avgProcessingTime != nil {
		stats["avg_processing_time_seconds"] = *avgProcessingTime
	}

	return stats, nil
}

func (db *Database) GetWalletPerformance(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
        SELECT 
            w.address,
            w.chain_id,
            COUNT(DISTINCT b.id) as total_batches,
            COUNT(DISTINCT bs.swap_id) as total_swaps,
            AVG(CAST(b.gas_price AS NUMERIC)) as avg_gas_price,
            SUM(b.gas_used) as total_gas_used
        FROM hot_wallets w
        LEFT JOIN batches b ON w.address = b.wallet_address
        LEFT JOIN batch_swaps bs ON b.id = bs.batch_id
        WHERE w.is_active = true
        GROUP BY w.address, w.chain_id
    `

	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting wallet performance: %v", err)
	}
	defer rows.Close()

	var performance []map[string]interface{}
	for rows.Next() {
		var wallet = make(map[string]interface{})
		var avgGasPrice, totalGasUsed *float64

		err := rows.Scan(
			&wallet["address"],
			&wallet["chain_id"],
			&wallet["total_batches"],
			&wallet["total_swaps"],
			&avgGasPrice,
			&totalGasUsed,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning wallet performance: %v", err)
		}

		if avgGasPrice != nil {
			wallet["avg_gas_price"] = *avgGasPrice
		}
		if totalGasUsed != nil {
			wallet["total_gas_used"] = *totalGasUsed
		}

		performance = append(performance, wallet)
	}

	return performance, nil
}
