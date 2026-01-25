-- Add deleted_remotely flag to planned_manufacturing_orders
ALTER TABLE planned_manufacturing_orders
ADD COLUMN deleted_remotely BOOLEAN NOT NULL DEFAULT false;

-- Add deleted_remotely flag to manufacturing_orders
ALTER TABLE manufacturing_orders
ADD COLUMN deleted_remotely BOOLEAN NOT NULL DEFAULT false;

-- Add deleted_remotely flag to production_orders (unified table)
ALTER TABLE production_orders
ADD COLUMN deleted_remotely BOOLEAN NOT NULL DEFAULT false;

-- Index for filtering in issue detection queries (partial index for efficiency)
CREATE INDEX idx_planned_mfg_orders_deleted_remotely
    ON planned_manufacturing_orders(deleted_remotely) WHERE deleted_remotely = true;

CREATE INDEX idx_mfg_orders_deleted_remotely
    ON manufacturing_orders(deleted_remotely) WHERE deleted_remotely = true;

CREATE INDEX idx_production_orders_deleted_remotely
    ON production_orders(deleted_remotely) WHERE deleted_remotely = true;
