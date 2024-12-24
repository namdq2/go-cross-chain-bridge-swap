package processor

import (
    "context"
    "crypto/ecdsa"
    "sync"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
    PrivateKey   *ecdsa.PrivateKey
    Address      common.Address
    NonceMap     map[int64]uint64
    IsProcessing bool
    LastUsed     time.Time
    mutex        sync.Mutex
}

type WalletPool struct {
    wallets []*Wallet
    mutex   sync.RWMutex
}

func NewWalletPool(privateKeys []string) (*WalletPool, error) {
    pool := &WalletPool{
        wallets: make([]*Wallet, len(privateKeys)),
    }

    for i, privKey := range privateKeys {
        key, err := crypto.HexToECDSA(privKey)
        if err != nil {
            return nil, err
        }

        wallet := &Wallet{
            PrivateKey: key,
            Address:    crypto.PubkeyToAddress(key.PublicKey),
            NonceMap:   make(map[int64]uint64),
            LastUsed:   time.Now(),
        }
        pool.wallets[i] = wallet
    }

    return pool, nil
}

func (wp *WalletPool) getAvailableWallet() *Wallet {
    wp.mutex.Lock()
    defer wp.mutex.Unlock()

    var selectedWallet *Wallet
    oldestLastUsed := time.Now()

    for _, wallet := range wp.wallets {
        wallet.mutex.Lock()
        if !wallet.IsProcessing && wallet.LastUsed.Before(oldestLastUsed) {
            selectedWallet = wallet
            oldestLastUsed = wallet.LastUsed
        }
        wallet.mutex.Unlock()
    }

    if selectedWallet != nil {
        selectedWallet.mutex.Lock()
        selectedWallet.IsProcessing = true
        selectedWallet.LastUsed = time.Now()
        selectedWallet.mutex.Unlock()
    }

    return selectedWallet
}

func (wp *WalletPool) releaseWallet(wallet *Wallet) {
    wallet.mutex.Lock()
    wallet.IsProcessing = false
    wallet.mutex.Unlock()
}

func (w *Wallet) ProcessBatch(chainID int64, batch []*models.SwapRequest, batchID int64) error {
    // Implementation for processing batch with this wallet
    return nil
}