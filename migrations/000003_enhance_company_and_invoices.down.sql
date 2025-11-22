-- Drop invoice line items
DROP INDEX IF EXISTS idx_invoice_line_items_invoice;
DROP TABLE IF EXISTS invoice_line_items;

-- Remove company fields
ALTER TABLE companies DROP COLUMN logo_path;
ALTER TABLE companies DROP COLUMN bank_swift;
ALTER TABLE companies DROP COLUMN bank_account;
ALTER TABLE companies DROP COLUMN bank_name;
ALTER TABLE companies DROP COLUMN registration_address;
ALTER TABLE companies DROP COLUMN phone;
