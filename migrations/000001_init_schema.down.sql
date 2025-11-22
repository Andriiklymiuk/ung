-- Drop indexes
DROP INDEX IF EXISTS idx_tracking_sessions_client;
DROP INDEX IF EXISTS idx_invoice_recipients_client;
DROP INDEX IF EXISTS idx_invoice_recipients_invoice;
DROP INDEX IF EXISTS idx_invoices_company;

-- Drop tables
DROP TABLE IF EXISTS tracking_sessions;
DROP TABLE IF EXISTS invoice_recipients;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS companies;
