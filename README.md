# Cross-Chain Bridge Swap

A high-performance bridge application for swapping tokens between Ethereum and BSC chains with batch processing and multi-wallet support.

## Features

### Core Features
- Cross-chain token swapping between Ethereum and BSC
- Multi-wallet support for parallel processing
- Batch processing for gas optimization
- Automatic nonce management
- Chain reorganization handling
- PostgreSQL storage for transaction tracking
- Comprehensive monitoring and analytics

### Technical Features
- Concurrent batch processing
- Automatic gas price management
- Transaction retry mechanism
- Chain reorg detection and handling
- Multiple hot wallet support
- Database-backed queue system
- Real-time status updates

### Security Features
- Multi-layer validation
- Signature verification
- Rate limiting
- Error handling
- Transaction monitoring
- Audit logging

## Architecture

### Components
- Smart Contracts: Token locking and unlocking
- Batch Processor: Groups transactions for efficiency
- Wallet Manager: Handles multiple hot wallets
- Queue System: Manages pending transactions
- Database: PostgreSQL for persistence
- API Server: RESTful endpoints

### Database Schema
- `swaps`: Individual swap requests
- `batches`: Grouped transactions
- `batch_swaps`: Mapping between batches and swaps
- `hot_wallets`: Wallet management
- `supported_tokens`: Token whitelist
- `chain_configs`: Chain-specific settings
- `audit_logs`: System audit trail

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose
- Node.js 18+ (for contract deployment)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/namdq2/go-cross-chain-bridge-swap.git
cd go-cross-chain-bridge-swap
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Install dependencies:
```bash
go mod download
```

4. Deploy smart contracts:
```bash
cd contracts
npm install
npx hardhat compile
npx hardhat deploy --network ethereum
npx hardhat deploy --network bsc
```

5. Initialize database:
```bash
psql -U your_user -d your_db -f database/schema.sql
```

## Docker Deployment

1. Build and start services:
```bash
docker-compose build
docker-compose up -d
```

2. Monitor logs:
```bash
docker-compose logs -f app
docker-compose logs -f postgres
```

3. Scale if needed:
```bash
docker-compose up -d --scale app=3
```

## Configuration

### Environment Variables

```env
# Chain RPC endpoints
CHAIN1_RPC=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
CHAIN2_RPC=https://bsc-dataseed.binance.org

# Smart contract addresses
BRIDGE_ADDR1=0x...  # Ethereum bridge contract
BRIDGE_ADDR2=0x...  # BSC bridge contract

# Hot wallet private keys (comma-separated)
HOT_WALLET_PRIVATE_KEYS=key1,key2,key3

# Database configuration
DATABASE_URL=postgresql://user:password@localhost:5432/bridge_db

# Optional configurations
MAX_BATCH_SIZE=50
BATCH_TIMEOUT=30s
MAX_GAS_PRICE_GWEI=500
```

### Chain Configurations
```sql
INSERT INTO chain_configs (
    chain_id, chain_type, rpc_url, 
    bridge_address, required_confirmations
) VALUES
(1, 'ethereum', 'https://mainnet.infura.io/v3/YOUR-KEY', '0x...', 12),
(56, 'bsc', 'https://bsc-dataseed.binance.org', '0x...', 20);
```

## API Documentation

### Initiate Swap
```http
POST /api/swap
Content-Type: application/json

{
    "fromChainId": 1,
    "toChainId": 56,
    "tokenAddress": "0x...",
    "amount": "1000000000000000000",
    "recipient": "0x..."
}
```

Response:
```json
{
    "status": "pending",
    "requestId": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Swap request has been queued for next batch",
    "maxBatchWait": 30
}
```

### Get Swap Status
```http
GET /api/swap/{requestId}
```

Response:
```json
{
    "requestId": "550e8400-e29b-41d4-a716-446655440000",
    "status": "completed",
    "fromChainId": 1,
    "toChainId": 56,
    "sourceTxHash": "0x...",
    "targetTxHash": "0x...",
    "timestamp": "2024-12-24T10:00:00Z"
}
```

### Get Queue Status
```http
GET /api/queue/status
```

Response:
```json
{
    "queueLength": 10,
    "maxQueueSize": 1000,
    "activeWorkers": 5,
    "processingBatches": 2
}
```

## Monitoring & Analytics

### Available Metrics
- Swap success/failure rates
- Average processing time
- Gas usage statistics
- Wallet performance
- Queue length and processing rate

### Database Views
- `swap_statistics`: Aggregated swap metrics
- `wallet_performance`: Hot wallet analytics

## Development

### Project Structure
```
bridge-swap/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── api/                    # API handlers
│   ├── models/                 # Database models
│   ├── processor/              # Batch processing
│   └── service/                # Business logic
├── contracts/                  # Smart contracts
├── database/                   # SQL schemas
├── scripts/                    # Utility scripts
└── config/                     # Configuration
```

### Running Tests
```bash
go test ./...
```

### Local Development
1. Start PostgreSQL:
```bash
docker-compose up postgres -d
```

2. Run the application:
```bash
go run cmd/main.go
```

## Production Deployment

### Requirements
- High-availability setup
- Load balancer
- Database replication
- Monitoring system
- Backup system

### Recommended Setup
1. Multiple application instances
2. PostgreSQL with replication
3. Redis for caching (optional)
4. Prometheus + Grafana for monitoring
5. Regular database backups

### Security Considerations
1. Secure private key management
2. Rate limiting
3. DDoS protection
4. Regular security audits
5. Access control

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, email namquoc.dev@gmail.com or create an issue in the GitHub repository.

## Authors

- Nam Dang (namquoc.dev@gmail.com)

## Acknowledgments

- OpenZeppelin for smart contract libraries
- Go-Ethereum team for the Ethereum client
- PostgreSQL team for the database

4YoyBu4Xa4fwc9FrnsxqT582ff4Cdqoh62FVDZAMpyZj