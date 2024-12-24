package processor

import (
    "context"
    "sync"
    "time"

    "github.com/yourusername/bridge-swap/internal/models"
)

const (
    BATCH_SIZE    = 50
    BATCH_TIMEOUT = 30 * time.Second
)

type BatchProcessor struct {
    currentBatch  []*models.SwapRequest
    batchMutex    sync.Mutex
    batchTimer    *time.Timer
    walletPool    *WalletPool
    db            *models.Database
    processChan   chan struct{}
}

func NewBatchProcessor(walletPool *WalletPool, db *models.Database) *BatchProcessor {
    bp := &BatchProcessor{
        walletPool:  walletPool,
        db:         db,
        processChan: make(chan struct{}, 1),
    }
    bp.startNewBatch()
    go bp.processLoop()
    return bp
}

func (bp *BatchProcessor) startNewBatch() {
    bp.batchMutex.Lock()
    defer bp.batchMutex.Unlock()

    bp.currentBatch = make([]*models.SwapRequest, 0, BATCH_SIZE)
    bp.batchTimer = time.NewTimer(BATCH_TIMEOUT)

    go func() {
        <-bp.batchTimer.C
        bp.triggerProcess()
    }()
}

func (bp *BatchProcessor) AddRequest(req *models.SwapRequest) {
    bp.batchMutex.Lock()
    defer bp.batchMutex.Unlock()

    bp.currentBatch = append(bp.currentBatch, req)

    if len(bp.currentBatch) >= BATCH_SIZE {
        bp.batchTimer.Stop()
        bp.triggerProcess()
    }
}

func (bp *BatchProcessor) triggerProcess() {
    select {
    case bp.processChan <- struct{}{}:
    default:
    }
}

func (bp *BatchProcessor) processLoop() {
    for range bp.processChan {
        bp.processBatch()
    }
}

func (bp *BatchProcessor) processBatch() {
    bp.batchMutex.Lock()
    if len(bp.currentBatch) == 0 {
        bp.startNewBatch()
        bp.batchMutex.Unlock()
        return
    }

    batch := bp.currentBatch
    bp.startNewBatch()
    bp.batchMutex.Unlock()

    // Group by chain
    chainBatches := make(map[int64][]*models.SwapRequest)
    for _, req := range batch {
        chainBatches[req.FromChainID] = append(chainBatches[req.FromChainID], req)
    }

    // Process each chain's batch
    var wg sync.WaitGroup
    for chainID, chainBatch := range chainBatches {
        wg.Add(1)
        go func(cid int64, cbatch []*models.SwapRequest) {
            defer wg.Done()
            
            wallet := bp.walletPool.getAvailableWallet()
            if wallet == nil {
                return
            }
            defer bp.walletPool.releaseWallet(wallet)

            if err := bp.processChainBatch(cid, cbatch, wallet); err != nil {
                // Handle error
            }
        }(chainID, chainBatch)
    }
    wg.Wait()
}

func (bp *BatchProcessor) processChainBatch(chainID int64, batch []*models.SwapRequest, wallet *Wallet) error {
    ctx := context.Background()
    
    // Create batch record
    batchRecord, err := bp.db.CreateBatch(ctx, wallet.Address.Hex(), chainID)
    if err != nil {
        return err
    }

    // Add swaps to batch
    swapIDs := make([]int64, len(batch))
    for i, swap := range batch {
        swapIDs[i] = swap.ID
    }
    
    if err := bp.db.AddSwapsToBatch(ctx, batchRecord.ID, swapIDs); err != nil {
        return err
    }

    // Process on chain
    return wallet.ProcessBatch(chainID, batch, batchRecord.ID)
}