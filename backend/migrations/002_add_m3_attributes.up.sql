-- Add M3 attributes and metadata to existing tables

-- ========================================
-- Customer Order Lines - Add M3 Fields
-- ========================================

ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS cono INTEGER;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS divi VARCHAR(10);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS posx VARCHAR(10);

-- Reference order fields (critical for linking to MOs/MOPs)
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rorc INTEGER;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rorn VARCHAR(50);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rorl INTEGER;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rorx INTEGER;

-- Additional quantities
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rnqt DECIMAL(15,6);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS alqt DECIMAL(15,6);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS ivqt DECIMAL(15,6);

-- Additional dates
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS dwdt DATE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS codt DATE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS pldt DATE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS fded DATE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS lded DATE;

-- Pricing fields
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS sapr DECIMAL(15,4);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS nepr DECIMAL(15,4);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS lnam DECIMAL(15,2);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS cucd VARCHAR(10);

-- Attribute model reference
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS atnr BIGINT;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS atmo VARCHAR(20);
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS atpr VARCHAR(20);

-- All attributes stored as JSONB for flexibility
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS attributes JSONB;

-- M3 metadata fields
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rgdt DATE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS rgtm INTEGER;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS lmdt DATE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS lmts BIGINT;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS chno INTEGER;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS chid VARCHAR(20);

-- Data Lake metadata
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS m3_timestamp TIMESTAMP;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE customer_order_lines ADD COLUMN IF NOT EXISTS sync_timestamp TIMESTAMP DEFAULT NOW();

-- Create indexes for CO lines
CREATE INDEX IF NOT EXISTS idx_co_lines_rorc_rorn ON customer_order_lines(rorc, rorn, rorl, rorx);
CREATE INDEX IF NOT EXISTS idx_co_lines_cono_divi ON customer_order_lines(cono, divi);
CREATE INDEX IF NOT EXISTS idx_co_lines_lmdt ON customer_order_lines(lmdt);
CREATE INDEX IF NOT EXISTS idx_co_lines_attributes ON customer_order_lines USING GIN(attributes);

-- ========================================
-- Manufacturing Orders - Add M3 Fields
-- ========================================

ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS cono INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS divi VARCHAR(10);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS mfno VARCHAR(50);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS prno VARCHAR(50);

-- Status fields
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS whhs VARCHAR(10);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS wmst VARCHAR(10);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS mohs VARCHAR(10);

-- Additional quantities
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS orqa DECIMAL(15,6);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rvqt DECIMAL(15,6);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rvqa DECIMAL(15,6);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS maqt DECIMAL(15,6);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS maqa DECIMAL(15,6);

-- Additional dates
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rsdt DATE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS refd DATE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rpdt DATE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS fstd DATE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS ffid DATE;

-- Planning fields
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS prio INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS plgr VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS wcln VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS prdy INTEGER;

-- Reference order fields (links to CO lines)
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rorc INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rorn VARCHAR(50);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rorl INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rorx INTEGER;

-- Hierarchy fields
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS prhl VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS mfhl VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS prlo VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS mflo VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS levl INTEGER;

-- Attribute references
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS atnr BIGINT;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS cfin BIGINT;

-- Project fields
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS proj VARCHAR(20);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS elno VARCHAR(20);

-- Additional metadata
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS orty VARCHAR(10);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS getp VARCHAR(10);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS bdcd VARCHAR(10);
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS scex VARCHAR(10);

-- All attributes as JSONB
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS attributes JSONB;

-- M3 metadata
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rgdt DATE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS rgtm INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS lmdt DATE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS lmts BIGINT;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS chno INTEGER;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS chid VARCHAR(20);

-- Data Lake metadata
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS m3_timestamp TIMESTAMP;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE manufacturing_orders ADD COLUMN IF NOT EXISTS sync_timestamp TIMESTAMP DEFAULT NOW();

-- Create indexes for MOs
CREATE INDEX IF NOT EXISTS idx_mo_rorc_rorn ON manufacturing_orders(rorc, rorn, rorl, rorx);
CREATE INDEX IF NOT EXISTS idx_mo_cono_divi ON manufacturing_orders(cono, divi);
CREATE INDEX IF NOT EXISTS idx_mo_mfno ON manufacturing_orders(facility, mfno);
CREATE INDEX IF NOT EXISTS idx_mo_prno ON manufacturing_orders(prno);
CREATE INDEX IF NOT EXISTS idx_mo_lmdt ON manufacturing_orders(lmdt);
CREATE INDEX IF NOT EXISTS idx_mo_whhs ON manufacturing_orders(whhs);
CREATE INDEX IF NOT EXISTS idx_mo_attributes ON manufacturing_orders USING GIN(attributes);

-- ========================================
-- Planned Manufacturing Orders - Add M3 Fields
-- ========================================

ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS cono INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS divi VARCHAR(10);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS plpn BIGINT;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS plps INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS prno VARCHAR(50);

-- Status fields
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS psts VARCHAR(10);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS whst VARCHAR(10);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS actp VARCHAR(10);

-- Additional quantities
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS ppqt DECIMAL(15,6);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS orqa DECIMAL(15,6);

-- Additional dates
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS reld DATE;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS msti DATE;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS mfti DATE;

-- Planning fields
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS prip VARCHAR(10);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS plgr VARCHAR(20);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS wcln VARCHAR(20);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS prdy INTEGER;

-- Reference order fields (links to CO lines)
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rorc INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rorn VARCHAR(50);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rorl INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rorx INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rorh INTEGER;

-- Hierarchy fields
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS pllo BIGINT;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS plhl BIGINT;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS levl INTEGER;

-- Attribute references
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS atnr BIGINT;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS cfin BIGINT;

-- Project fields
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS proj VARCHAR(20);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS elno VARCHAR(20);

-- Additional metadata
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS gety VARCHAR(10);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS nuau INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS ordp VARCHAR(10);
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS orty VARCHAR(10);

-- Warning messages as JSONB
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS messages JSONB;

-- All attributes as JSONB
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS attributes JSONB;

-- M3 metadata
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rgdt DATE;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS rgtm INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS lmdt DATE;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS lmts BIGINT;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS chno INTEGER;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS chid VARCHAR(20);

-- Data Lake metadata
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS m3_timestamp TIMESTAMP;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE planned_manufacturing_orders ADD COLUMN IF NOT EXISTS sync_timestamp TIMESTAMP DEFAULT NOW();

-- Create indexes for MOPs
CREATE INDEX IF NOT EXISTS idx_mop_rorc_rorn ON planned_manufacturing_orders(rorc, rorn, rorl, rorx);
CREATE INDEX IF NOT EXISTS idx_mop_cono_divi ON planned_manufacturing_orders(cono, divi);
CREATE INDEX IF NOT EXISTS idx_mop_plpn ON planned_manufacturing_orders(facility, plpn, plps);
CREATE INDEX IF NOT EXISTS idx_mop_prno ON planned_manufacturing_orders(prno);
CREATE INDEX IF NOT EXISTS idx_mop_lmdt ON planned_manufacturing_orders(lmdt);
CREATE INDEX IF NOT EXISTS idx_mop_psts ON planned_manufacturing_orders(psts);
CREATE INDEX IF NOT EXISTS idx_mop_attributes ON planned_manufacturing_orders USING GIN(attributes);

-- ========================================
-- Production Orders - Update with linking fields
-- ========================================

ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS cono INTEGER;
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS divi VARCHAR(10);

-- Reference order fields for linking
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS rorc INTEGER;
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS rorn VARCHAR(50);
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS rorl INTEGER;
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS rorx INTEGER;

-- M3 metadata
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS lmdt DATE;
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS lmts BIGINT;
ALTER TABLE production_orders ADD COLUMN IF NOT EXISTS sync_timestamp TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_prod_orders_rorc_rorn ON production_orders(rorc, rorn, rorl, rorx);
CREATE INDEX IF NOT EXISTS idx_prod_orders_lmdt ON production_orders(lmdt);

-- ========================================
-- Customer Orders - Add M3 Fields
-- ========================================

ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS cono INTEGER;
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS divi VARCHAR(10);
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS orno VARCHAR(50);
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS cuno VARCHAR(50);

-- M3 metadata
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS rgdt DATE;
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS lmdt DATE;
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS lmts BIGINT;
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS m3_timestamp TIMESTAMP;
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE customer_orders ADD COLUMN IF NOT EXISTS sync_timestamp TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_co_cono_divi ON customer_orders(cono, divi);
CREATE INDEX IF NOT EXISTS idx_co_orno ON customer_orders(orno);
CREATE INDEX IF NOT EXISTS idx_co_lmdt ON customer_orders(lmdt);

-- ========================================
-- Deliveries - Add M3 Fields
-- ========================================

ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS cono INTEGER;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS divi VARCHAR(10);

-- M3 metadata
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS lmdt DATE;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS lmts BIGINT;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS m3_timestamp TIMESTAMP;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS sync_timestamp TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_deliveries_lmdt ON deliveries(lmdt);

-- ========================================
-- Comments on JSONB Attributes Structure
-- ========================================

COMMENT ON COLUMN customer_order_lines.attributes IS
'JSONB structure: {
  "builtin_numeric": {"ATV1": 123.45, "ATV2": 67.89, ...},
  "builtin_string": {"ATV6": "Color:Red", "ATV7": "Size:Large", ...},
  "user_defined_alpha": {"UCA1": "Value1", "UCA2": "Value2", ...},
  "user_defined_numeric": {"UDN1": 1000.50, "UDN2": 2500.75, ...},
  "user_defined_dates": {"UID1": "2024-01-15", "UID2": "2024-02-20", ...},
  "discounts": {"DIP1": 5.0, "DIP2": 2.5, "DIA1": 100.00}
}';

COMMENT ON COLUMN manufacturing_orders.attributes IS
'JSONB structure for MO-specific attributes and custom fields';

COMMENT ON COLUMN planned_manufacturing_orders.attributes IS
'JSONB structure for MOP-specific attributes and custom fields';

COMMENT ON COLUMN planned_manufacturing_orders.messages IS
'JSONB structure: {"MSG1": "message1", "MSG2": "message2", "MSG3": "message3", "MSG4": "message4"}';
