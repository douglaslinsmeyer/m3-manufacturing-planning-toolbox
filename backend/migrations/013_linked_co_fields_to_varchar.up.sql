-- Migration 013: Change linked_co fields to VARCHAR to match Data Fabric source format
-- Data Fabric returns MPREAL.DRDL and DRDX as strings, so store them as strings

-- Manufacturing Orders
ALTER TABLE manufacturing_orders
    ALTER COLUMN linked_co_line TYPE VARCHAR(10),
    ALTER COLUMN linked_co_suffix TYPE VARCHAR(10);

-- Planned Manufacturing Orders
ALTER TABLE planned_manufacturing_orders
    ALTER COLUMN linked_co_line TYPE VARCHAR(10),
    ALTER COLUMN linked_co_suffix TYPE VARCHAR(10);
