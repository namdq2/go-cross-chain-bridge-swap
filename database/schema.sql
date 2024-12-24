-- database/schema.sql

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enums
CREATE TYPE swap_status AS ENUM (
    'pending',
    'queued',
    'processing',
    'completed',
    'failed',
    'reverted'
);

CREATE TYPE chain_type AS ENUM (
    'ethereum',
    'bsc'
);

-- Create swaps table
CREATE TABLE swaps (
    id SERIAL PRIMARY KEY,
    request_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    from_chain_id BIGINT NOT NULL,
    to_chain_id BIGINT NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    amount NUMERIC(78) NOT NULL, -- To handle large token amounts
    recipient VARCHAR(42) NOT NULL,
    status swap_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create batches table
CREATE TABLE batches (
    id SERIAL PRIMARY KEY,
    batch_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    chain_id BIGINT NOT NULL,
    source_tx_hash VARCHAR(66),
    target_tx_hash VARCHAR(66),
    status swap_status NOT NULL DEFAULT 'pending',
    gas_price NUMERIC(78),
    gas_used BIGINT,
    block_number BIGINT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create batch_swaps mapping table
CREATE TABLE batch_swaps (
    batch_id INTEGER REFERENCES batches(id) ON DELETE CASCADE,
    swap_id INTEGER REFERENCES swaps(id) ON DELETE CASCADE,
    PRIMARY KEY (batch_id, swap_id)
);

-- Create hot wallets table
CREATE TABLE hot_wallets (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) UNIQUE NOT NULL,
    chain_id BIGINT NOT NULL,
    nonce BIGINT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    total_processed_batches INTEGER NOT NULL DEFAULT 0,
    total_processed_volume NUMERIC(78) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create supported tokens table
CREATE TABLE supported_tokens (
    id SERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    token_symbol VARCHAR(10) NOT NULL,
    token_decimals INTEGER NOT NULL,
    min_amount NUMERIC(78) NOT NULL,
    max_amount NUMERIC(78) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(chain_id, token_address)
);

-- Create chain_configs table
CREATE TABLE chain_configs (
    id SERIAL PRIMARY KEY,
    chain_id BIGINT UNIQUE NOT NULL,
    chain_type chain_type NOT NULL,
    rpc_url TEXT NOT NULL,
    bridge_address VARCHAR(42) NOT NULL,
    required_confirmations INTEGER NOT NULL,
    max_gas_price NUMERIC(78),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create audit_logs table
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_swaps_status ON swaps(status);
CREATE INDEX idx_swaps_request_id ON swaps(request_id);
CREATE INDEX idx_swaps_from_chain ON swaps(from_chain_id);
CREATE INDEX idx_swaps_to_chain ON swaps(to_chain_id);
CREATE INDEX idx_swaps_token ON swaps(token_address);
CREATE INDEX idx_swaps_recipient ON swaps(recipient);
CREATE INDEX idx_swaps_created_at ON swaps(created_at);

CREATE INDEX idx_batches_status ON batches(status);
CREATE INDEX idx_batches_wallet ON batches(wallet_address);
CREATE INDEX idx_batches_chain ON batches(chain_id);
CREATE INDEX idx_batches_created_at ON batches(created_at);
CREATE INDEX idx_batches_block_number ON batches(block_number);

CREATE INDEX idx_hot_wallets_chain ON hot_wallets(chain_id);
CREATE INDEX idx_hot_wallets_active ON hot_wallets(is_active);
CREATE INDEX idx_hot_wallets_last_used ON hot_wallets(last_used_at);

CREATE INDEX idx_supported_tokens_chain ON supported_tokens(chain_id);
CREATE INDEX idx_supported_tokens_active ON supported_tokens(is_active);

-- Updated timestamp triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_swaps_updated_at
    BEFORE UPDATE ON swaps
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_batches_updated_at
    BEFORE UPDATE ON batches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hot_wallets_updated_at
    BEFORE UPDATE ON hot_wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_supported_tokens_updated_at
    BEFORE UPDATE ON supported_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chain_configs_updated_at
    BEFORE UPDATE ON chain_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Audit logging trigger
CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (
            entity_type,
            entity_id,
            action,
            new_data
        ) VALUES (
            TG_TABLE_NAME,
            NEW.id,
            'INSERT',
            row_to_json(NEW)
        );
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs (
            entity_type,
            entity_id,
            action,
            old_data,
            new_data
        ) VALUES (
            TG_TABLE_NAME,
            NEW.id,
            'UPDATE',
            row_to_json(OLD),
            row_to_json(NEW)
        );
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs (
            entity_type,
            entity_id,
            action,
            old_data
        ) VALUES (
            TG_TABLE_NAME,
            OLD.id,
            'DELETE',
            row_to_json(OLD)
        );
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_swaps_trigger
    AFTER INSERT OR UPDATE OR DELETE ON swaps
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER audit_batches_trigger
    AFTER INSERT OR UPDATE OR DELETE ON batches
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER audit_hot_wallets_trigger
    AFTER INSERT OR UPDATE OR DELETE ON hot_wallets
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

-- Views for analytics
CREATE OR REPLACE VIEW swap_statistics AS
SELECT
    date_trunc('hour', created_at) as time_bucket,
    from_chain_id,
    to_chain_id,
    token_address,
    count(*) as total_swaps,
    sum(amount) as total_volume,
    count(CASE WHEN status = 'completed' THEN 1 END) as successful_swaps,
    count(CASE WHEN status = 'failed' THEN 1 END) as failed_swaps,
    avg(EXTRACT(EPOCH FROM (updated_at - created_at))) as avg_processing_time
FROM swaps
GROUP BY time_bucket, from_chain_id, to_chain_id, token_address;

CREATE OR REPLACE VIEW wallet_performance AS
SELECT
    w.address,
    w.chain_id,
    count(DISTINCT b.id) as total_batches,
    count(DISTINCT bs.swap_id) as total_swaps,
    avg(b.gas_price) as avg_gas_price,
    sum(b.gas_used) as total_gas_used,
    count(CASE WHEN b.status = 'completed' THEN 1 END) as successful_batches,
    count(CASE WHEN b.status = 'failed' THEN 1 END) as failed_batches
FROM hot_wallets w
LEFT JOIN batches b ON w.address = b.wallet_address
LEFT JOIN batch_swaps bs ON b.id = bs.batch_id
GROUP BY w.address, w.chain_id;

-- Initial data
INSERT INTO chain_configs (chain_id, chain_type, rpc_url, bridge_address, required_confirmations, max_gas_price) VALUES
(1, 'ethereum', 'https://mainnet.infura.io/v3/YOUR-PROJECT-ID', '0x...', 12, 500000000000),
(56, 'bsc', 'https://bsc-dataseed.binance.org', '0x...', 20, 10000000000);