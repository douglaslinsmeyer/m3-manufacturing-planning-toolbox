-- Migration 013 DOWN: Revert linked_co fields back to INTEGER

-- Manufacturing Orders
ALTER TABLE manufacturing_orders
    ALTER COLUMN linked_co_line TYPE INTEGER USING linked_co_line::INTEGER,
    ALTER COLUMN linked_co_suffix TYPE INTEGER USING linked_co_suffix::INTEGER;

-- Planned Manufacturing Orders
ALTER TABLE planned_manufacturing_orders
    ALTER COLUMN linked_co_line TYPE INTEGER USING linked_co_line::INTEGER,
    ALTER COLUMN linked_co_suffix TYPE INTEGER USING linked_co_suffix::INTEGER;
