-- Migration 018 Down: Remove linked CO fields from production_orders

DROP INDEX IF EXISTS idx_prod_linked_co;

ALTER TABLE production_orders
    DROP COLUMN IF EXISTS linked_co_number,
    DROP COLUMN IF EXISTS linked_co_line,
    DROP COLUMN IF EXISTS linked_co_suffix,
    DROP COLUMN IF EXISTS allocated_qty;
