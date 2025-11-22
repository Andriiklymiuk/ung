-- Add additional company fields
ALTER TABLE companies ADD COLUMN phone TEXT;
ALTER TABLE companies ADD COLUMN registration_address TEXT;
ALTER TABLE companies ADD COLUMN bank_name TEXT;
ALTER TABLE companies ADD COLUMN bank_account TEXT;
ALTER TABLE companies ADD COLUMN bank_swift TEXT;
ALTER TABLE companies ADD COLUMN logo_path TEXT;

-- Create invoice line items table
CREATE TABLE IF NOT EXISTS invoice_line_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL,
    item_name TEXT NOT NULL,
    description TEXT,
    quantity REAL NOT NULL DEFAULT 1,
    rate REAL NOT NULL,
    amount REAL NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id) REFERENCES invoices(id)
);

CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice ON invoice_line_items(invoice_id);
