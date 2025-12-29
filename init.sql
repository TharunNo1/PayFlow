CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    owner_name TEXT NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE entries (
    id UUID PRIMARY KEY,
    account_id UUID REFERENCES accounts(id),
    amount BIGINT NOT NULL, -- Negative for debit, positive for credit
    transaction_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE payout_tasks (
    id UUID PRIMARY KEY,
    entry_id UUID REFERENCES entries(id),
    status TEXT NOT NULL DEFAULT 'PENDING', -- PENDING, PROCESSING, COMPLETED, FAILED
    retry_count INT DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);