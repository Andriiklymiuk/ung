-- Create contracts table
CREATE TABLE IF NOT EXISTS contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    contract_type TEXT NOT NULL,
    hourly_rate REAL,
    fixed_price REAL,
    currency TEXT DEFAULT 'USD',
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    active BOOLEAN DEFAULT 1,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

-- Add contract_id and hours to tracking_sessions
ALTER TABLE tracking_sessions ADD COLUMN contract_id INTEGER REFERENCES contracts(id);
ALTER TABLE tracking_sessions ADD COLUMN hours REAL;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_contracts_client ON contracts(client_id);
CREATE INDEX IF NOT EXISTS idx_contracts_active ON contracts(active);
CREATE INDEX IF NOT EXISTS idx_tracking_sessions_contract ON tracking_sessions(contract_id);
