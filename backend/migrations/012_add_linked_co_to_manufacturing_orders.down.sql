-- Migration 012 DOWN: Remove linked_co columns from manufacturing_orders

DROP INDEX IF EXISTS idx_mo_linked_co;

ALTER TABLE manufacturing_orders
    DROP COLUMN IF EXISTS linked_co_number,
    DROP COLUMN IF EXISTS linked_co_line,
    DROP COLUMN IF EXISTS linked_co_suffix,
    DROP COLUMN IF EXISTS allocated_qty;
