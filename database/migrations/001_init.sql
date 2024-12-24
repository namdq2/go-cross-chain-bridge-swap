CREATE TYPE swap_status AS ENUM ('pending', 'queued', 'processing', 'completed', 'failed');

CREATE TABLE swaps (
    id SERIAL PRIMARY KEY,
    request_id VARCHAR(36) UNIQUE NOT NULL,
    from_chain_id BIGINT NOT NULL,
    to_chain_id BIGINT NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    amount VARCHAR NOT NULL,
    recipient VARCHAR(42) NOT NULL,
    status swap_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE batches (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(66) UNIQUE NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    chain_id BIGINT NOT NULL,
    source_tx_hash VARCHAR(66),
    target_tx_hash VARCHAR(66),
    status swap_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE batch_swaps (
    batch_id INTEGER REFERENCES batches(id),
    swap_id INTEGER REFERENCES swaps(id),
    PRIMARY KEY (batch_id, swap_id)
);

CREATE INDEX idx_swaps_status ON swaps(status);
CREATE INDEX idx_batches_status ON batches(status);
CREATE INDEX idx_swaps_request_id ON swaps(request_id);