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

INSERT INTO accounts (id, owner_name, balance) VALUES 
('00000000-0000-0000-0000-000000000001', 'Company Payroll', 1000000),
('00000000-0000-0000-0000-000000000002', 'Employee One', 0)
ON CONFLICT (id) DO NOTHING;