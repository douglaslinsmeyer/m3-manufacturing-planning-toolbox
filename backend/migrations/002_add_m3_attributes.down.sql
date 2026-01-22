-- Rollback M3 attributes migration

-- Drop indexes
DROP INDEX IF EXISTS idx_co_lines_rorc_rorn;
DROP INDEX IF EXISTS idx_co_lines_cono_divi;
DROP INDEX IF EXISTS idx_co_lines_lmdt;
DROP INDEX IF EXISTS idx_co_lines_attributes;

DROP INDEX IF EXISTS idx_mo_rorc_rorn;
DROP INDEX IF EXISTS idx_mo_cono_divi;
DROP INDEX IF EXISTS idx_mo_mfno;
DROP INDEX IF EXISTS idx_mo_prno;
DROP INDEX IF EXISTS idx_mo_lmdt;
DROP INDEX IF EXISTS idx_mo_whhs;
DROP INDEX IF EXISTS idx_mo_attributes;

DROP INDEX IF EXISTS idx_mop_rorc_rorn;
DROP INDEX IF EXISTS idx_mop_cono_divi;
DROP INDEX IF EXISTS idx_mop_plpn;
DROP INDEX IF EXISTS idx_mop_prno;
DROP INDEX IF EXISTS idx_mop_lmdt;
DROP INDEX IF EXISTS idx_mop_psts;
DROP INDEX IF EXISTS idx_mop_attributes;

DROP INDEX IF EXISTS idx_prod_orders_rorc_rorn;
DROP INDEX IF EXISTS idx_prod_orders_lmdt;

DROP INDEX IF EXISTS idx_co_cono_divi;
DROP INDEX IF EXISTS idx_co_orno;
DROP INDEX IF EXISTS idx_co_lmdt;

DROP INDEX IF EXISTS idx_deliveries_lmdt;

-- Customer Order Lines
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS cono;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS divi;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS posx;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rorc;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rorn;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rorl;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rorx;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rnqt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS alqt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS ivqt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS dwdt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS codt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS pldt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS fded;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS lded;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS sapr;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS nepr;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS lnam;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS cucd;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS atnr;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS atmo;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS atpr;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS attributes;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rgdt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS rgtm;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS lmdt;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS lmts;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS chno;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS chid;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS m3_timestamp;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE customer_order_lines DROP COLUMN IF EXISTS sync_timestamp;

-- Manufacturing Orders
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS cono;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS divi;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS mfno;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS prno;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS whhs;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS wmst;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS mohs;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS orqa;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rvqt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rvqa;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS maqt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS maqa;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rsdt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS refd;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rpdt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS fstd;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS ffid;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS prio;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS plgr;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS wcln;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS prdy;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rorc;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rorn;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rorl;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rorx;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS prhl;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS mfhl;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS prlo;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS mflo;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS levl;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS atnr;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS cfin;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS proj;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS elno;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS orty;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS getp;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS bdcd;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS scex;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS attributes;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rgdt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS rgtm;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS lmdt;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS lmts;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS chno;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS chid;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS m3_timestamp;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE manufacturing_orders DROP COLUMN IF EXISTS sync_timestamp;

-- Planned Manufacturing Orders
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS cono;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS divi;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS plpn;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS plps;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS prno;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS psts;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS whst;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS actp;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS ppqt;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS orqa;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS reld;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS msti;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS mfti;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS prip;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS plgr;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS wcln;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS prdy;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rorc;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rorn;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rorl;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rorx;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rorh;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS pllo;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS plhl;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS levl;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS atnr;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS cfin;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS proj;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS elno;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS gety;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS nuau;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS ordp;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS orty;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS messages;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS attributes;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rgdt;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS rgtm;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS lmdt;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS lmts;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS chno;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS chid;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS m3_timestamp;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE planned_manufacturing_orders DROP COLUMN IF EXISTS sync_timestamp;

-- Production Orders
ALTER TABLE production_orders DROP COLUMN IF EXISTS cono;
ALTER TABLE production_orders DROP COLUMN IF EXISTS divi;
ALTER TABLE production_orders DROP COLUMN IF EXISTS rorc;
ALTER TABLE production_orders DROP COLUMN IF EXISTS rorn;
ALTER TABLE production_orders DROP COLUMN IF EXISTS rorl;
ALTER TABLE production_orders DROP COLUMN IF EXISTS rorx;
ALTER TABLE production_orders DROP COLUMN IF EXISTS lmdt;
ALTER TABLE production_orders DROP COLUMN IF EXISTS lmts;
ALTER TABLE production_orders DROP COLUMN IF EXISTS sync_timestamp;

-- Customer Orders
ALTER TABLE customer_orders DROP COLUMN IF EXISTS cono;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS divi;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS orno;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS cuno;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS rgdt;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS lmdt;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS lmts;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS m3_timestamp;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE customer_orders DROP COLUMN IF EXISTS sync_timestamp;

-- Deliveries
ALTER TABLE deliveries DROP COLUMN IF EXISTS cono;
ALTER TABLE deliveries DROP COLUMN IF EXISTS divi;
ALTER TABLE deliveries DROP COLUMN IF EXISTS lmdt;
ALTER TABLE deliveries DROP COLUMN IF EXISTS lmts;
ALTER TABLE deliveries DROP COLUMN IF EXISTS m3_timestamp;
ALTER TABLE deliveries DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE deliveries DROP COLUMN IF EXISTS sync_timestamp;
