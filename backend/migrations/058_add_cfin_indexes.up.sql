-- Add indexes on CFIN columns for better lookup performance
-- CFIN is used for historical CO traceability when MPREAL links are broken

-- Index on manufacturing_orders.cfin
-- Used by: unlinked concentration detector, CFIN lookup queries
CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_cfin
ON manufacturing_orders(cfin)
WHERE cfin IS NOT NULL AND cfin != '';

-- Index on planned_manufacturing_orders.cfin
-- Used by: unlinked concentration detector, CFIN lookup queries
CREATE INDEX IF NOT EXISTS idx_planned_manufacturing_orders_cfin
ON planned_manufacturing_orders(cfin)
WHERE cfin IS NOT NULL AND cfin != '';

-- Index on customer_order_lines.cfin
-- Used by: CFIN to CO lookup queries, concentration analysis
CREATE INDEX IF NOT EXISTS idx_customer_order_lines_cfin
ON customer_order_lines(cfin)
WHERE cfin IS NOT NULL AND cfin != '';

-- Composite index for unlinked concentration detector (CFIN + environment + status)
CREATE INDEX IF NOT EXISTS idx_planned_mfg_orders_cfin_unlinked
ON planned_manufacturing_orders(cfin, environment, whlo)
WHERE (linked_co_number IS NULL OR linked_co_number = '')
  AND deleted_remotely = false
  AND psts = '20'
  AND cfin IS NOT NULL
  AND cfin != '';
