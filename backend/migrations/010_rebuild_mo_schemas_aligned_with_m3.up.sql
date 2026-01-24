-- Migration 010: Rebuild manufacturing order schemas aligned with M3 data structures
-- This migration drops and recreates all three tables to precisely match M3 MWOHED and MMOPLP tables

-- Drop existing tables (CASCADE to handle dependencies)
DROP TABLE IF EXISTS production_orders CASCADE;
DROP TABLE IF EXISTS manufacturing_orders CASCADE;
DROP TABLE IF EXISTS planned_manufacturing_orders CASCADE;

-- ============================================================================
-- Manufacturing Orders Table (Based on M3 MWOHED structure)
-- ============================================================================

CREATE TABLE manufacturing_orders (
    id BIGSERIAL PRIMARY KEY,

    -- M3 Core Identifiers
    cono INTEGER NOT NULL,
    divi VARCHAR(10),
    facility VARCHAR(10) NOT NULL,
    mo_number VARCHAR(50) NOT NULL,
    product_number VARCHAR(50),
    item_number VARCHAR(50) NOT NULL,

    -- M3 Status Fields
    whst VARCHAR(10),        -- MO status (main)
    whhs VARCHAR(10),        -- Warehouse status
    wmst VARCHAR(10),        -- Material status
    mohs VARCHAR(10),        -- Order head status

    -- M3 Quantities
    orqt DECIMAL(15,6),      -- Ordered quantity
    maqt DECIMAL(15,6),      -- Manufactured quantity
    orqa DECIMAL(15,6),      -- Ordered quantity alternative unit
    rvqt DECIMAL(15,6),      -- Received quantity
    rvqa DECIMAL(15,6),      -- Received quantity alternative
    maqa DECIMAL(15,6),      -- Manufactured quantity alternative

    -- M3 Date Fields (INTEGER YYYYMMDD format)
    stdt INTEGER,            -- Planned start date
    fidt INTEGER,            -- Planned finish date
    msti INTEGER,            -- Start date - material
    mfti INTEGER,            -- Finish date - material
    fstd INTEGER,            -- Actual start date
    ffid INTEGER,            -- Actual finish date
    rsdt INTEGER,            -- Rescheduled start date
    refd INTEGER,            -- Rescheduled finish date
    rpdt INTEGER,            -- Reported date

    -- M3 Planning Fields
    prio INTEGER,            -- Priority
    resp VARCHAR(50),        -- Responsible/Operator
    plgr VARCHAR(50),        -- Planner group
    wcln VARCHAR(50),        -- Production line
    prdy INTEGER,            -- Production days

    -- M3 Warehouse/Location
    whlo VARCHAR(10),        -- Warehouse
    whsl VARCHAR(50),        -- Location
    bano VARCHAR(50),        -- Batch/Lot number

    -- M3 Reference Orders
    rorc INTEGER,            -- Reference order category
    rorn VARCHAR(50),        -- Reference order number
    rorl INTEGER,            -- Reference order line
    rorx INTEGER,            -- Reference order line suffix

    -- M3 Hierarchy
    prhl VARCHAR(50),        -- Product structure highest level
    mfhl VARCHAR(50),        -- MO highest level
    prlo VARCHAR(50),        -- Product structure lowest level
    mflo VARCHAR(50),        -- MO lowest level
    levl INTEGER,            -- Structure level

    -- M3 Configuration/Attributes
    cfin BIGINT,             -- Configuration number
    atnr BIGINT,             -- Attribute number

    -- M3 Order Type/Origin
    orty VARCHAR(10),        -- Order type
    getp VARCHAR(10),        -- Origin type

    -- M3 Material/BOM
    bdcd VARCHAR(10),        -- BOM change date control
    scex VARCHAR(10),        -- Supply chain exception
    strt VARCHAR(10),        -- Product structure type
    ecve VARCHAR(50),        -- Engineering change version

    -- M3 Routing
    aoid VARCHAR(50),        -- Alternate routing ID
    nuop INTEGER,            -- Number of operations
    nufo INTEGER,            -- Number of final operations

    -- M3 Action/Text
    actp VARCHAR(10),        -- Action message type
    txt1 TEXT,               -- Text line 1
    txt2 TEXT,               -- Text line 2

    -- M3 Project
    proj VARCHAR(50),        -- Project number
    elno VARCHAR(50),        -- Element number

    -- M3 Audit Fields
    rgdt INTEGER,            -- Entry date
    rgtm INTEGER,            -- Entry time
    lmdt INTEGER,            -- Change date (YYYYMMDD)
    lmts BIGINT,             -- Change timestamp
    chno INTEGER,            -- Change number
    chid VARCHAR(50),        -- Changed by

    -- Data Lake Metadata
    m3_timestamp BIGINT,     -- M3 timestamp field
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Application Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT unique_mo_number UNIQUE (mo_number)
);

-- Indexes for manufacturing_orders
CREATE INDEX idx_mo_facility ON manufacturing_orders(facility);
CREATE INDEX idx_mo_item ON manufacturing_orders(item_number);
CREATE INDEX idx_mo_status ON manufacturing_orders(whst);
CREATE INDEX idx_mo_dates ON manufacturing_orders(stdt, fidt);
CREATE INDEX idx_mo_lmdt ON manufacturing_orders(lmdt);
CREATE INDEX idx_mo_ref_order ON manufacturing_orders(rorc, rorn, rorl);
CREATE INDEX idx_mo_warehouse ON manufacturing_orders(whlo);
CREATE INDEX idx_mo_planner ON manufacturing_orders(plgr);

-- ============================================================================
-- Planned Manufacturing Orders Table (Based on M3 MMOPLP structure)
-- ============================================================================

CREATE TABLE planned_manufacturing_orders (
    id BIGSERIAL PRIMARY KEY,

    -- M3 Core Identifiers
    cono INTEGER NOT NULL,
    divi VARCHAR(10),
    facility VARCHAR(10) NOT NULL,
    plpn BIGINT NOT NULL,          -- Planned order number
    plps INTEGER,                   -- Planned order sequence
    product_number VARCHAR(50),     -- PRNO
    item_number VARCHAR(50) NOT NULL,

    -- M3 Status Fields
    psts VARCHAR(10),        -- Planned order status (main)
    whst VARCHAR(10),        -- MO status
    actp VARCHAR(10),        -- Action message type

    -- M3 Order Type
    orty VARCHAR(10),        -- Order type
    gety VARCHAR(10),        -- Origin type

    -- M3 Quantities
    ppqt DECIMAL(15,6),      -- Planned quantity
    orqa DECIMAL(15,6),      -- Ordered quantity alternative

    -- M3 Date Fields (INTEGER YYYYMMDD format)
    reld INTEGER,            -- Release date
    stdt INTEGER,            -- Planned start date
    fidt INTEGER,            -- Planned finish date
    msti INTEGER,            -- Start date - material
    mfti INTEGER,            -- Finish date - material
    pldt INTEGER,            -- Planning date

    -- M3 Planning Fields
    resp VARCHAR(50),        -- Responsible
    prip INTEGER,            -- Priority
    plgr VARCHAR(50),        -- Planner group
    wcln VARCHAR(50),        -- Production line
    prdy INTEGER,            -- Production days

    -- M3 Warehouse
    whlo VARCHAR(10),        -- Warehouse

    -- M3 Reference Orders
    rorc INTEGER,            -- Reference order category
    rorn VARCHAR(50),        -- Reference order number
    rorl INTEGER,            -- Reference order line
    rorx INTEGER,            -- Reference order line suffix
    rorh VARCHAR(50),        -- Reference order header

    -- M3 Hierarchy
    pllo VARCHAR(50),        -- Planned order lowest level
    plhl VARCHAR(50),        -- Planned order highest level

    -- M3 Configuration/Attributes
    atnr BIGINT,             -- Attribute number
    cfin BIGINT,             -- Configuration number

    -- M3 Project
    proj VARCHAR(50),        -- Project number
    elno VARCHAR(50),        -- Element number

    -- M3 Messages (MSG1-MSG4 stored as JSONB)
    messages JSONB,

    -- M3 Planning Parameters
    nuau INTEGER,            -- Number auto-generated
    ordp VARCHAR(10),        -- Order priority

    -- M3 Audit Fields
    rgdt INTEGER,            -- Entry date
    rgtm INTEGER,            -- Entry time
    lmdt INTEGER,            -- Change date (YYYYMMDD)
    lmts BIGINT,             -- Change timestamp
    chno INTEGER,            -- Change number
    chid VARCHAR(50),        -- Changed by

    -- Data Lake Metadata
    m3_timestamp BIGINT,
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Customer Order Link (from MPREAL join)
    linked_co_number VARCHAR(50),
    linked_co_line INTEGER,
    linked_co_suffix INTEGER,
    allocated_qty DECIMAL(15,6),

    -- Application Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT unique_plpn UNIQUE (plpn)
);

-- Indexes for planned_manufacturing_orders
CREATE INDEX idx_mop_facility ON planned_manufacturing_orders(facility);
CREATE INDEX idx_mop_item ON planned_manufacturing_orders(item_number);
CREATE INDEX idx_mop_status ON planned_manufacturing_orders(psts);
CREATE INDEX idx_mop_dates ON planned_manufacturing_orders(stdt, fidt);
CREATE INDEX idx_mop_lmdt ON planned_manufacturing_orders(lmdt);
CREATE INDEX idx_mop_linked_co ON planned_manufacturing_orders(linked_co_number, linked_co_line);
CREATE INDEX idx_mop_warehouse ON planned_manufacturing_orders(whlo);
CREATE INDEX idx_mop_planner ON planned_manufacturing_orders(plgr);

-- ============================================================================
-- Production Orders Table (Unified View of MOs and MOPs)
-- ============================================================================

CREATE TABLE production_orders (
    id BIGSERIAL PRIMARY KEY,

    -- Discriminator
    order_type VARCHAR(10) NOT NULL,  -- 'MO' or 'MOP'
    order_number VARCHAR(50) NOT NULL,

    -- M3 Context
    cono INTEGER NOT NULL,
    divi VARCHAR(10),
    facility VARCHAR(10) NOT NULL,

    -- Item Information
    product_number VARCHAR(50),
    item_number VARCHAR(50) NOT NULL,

    -- Quantities
    ordered_quantity DECIMAL(15,6),
    manufactured_quantity DECIMAL(15,6),

    -- Dates (INTEGER YYYYMMDD)
    planned_start_date INTEGER,
    planned_finish_date INTEGER,
    actual_start_date INTEGER,
    actual_finish_date INTEGER,
    release_date INTEGER,
    material_start_date INTEGER,
    material_finish_date INTEGER,

    -- Status
    status VARCHAR(10),
    proposal_status VARCHAR(10),

    -- Planning
    priority INTEGER,
    responsible VARCHAR(50),
    planner_group VARCHAR(50),
    production_line VARCHAR(50),

    -- Warehouse/Location
    warehouse VARCHAR(10),
    location VARCHAR(50),
    batch_number VARCHAR(50),

    -- Reference Orders
    rorc INTEGER,
    rorn VARCHAR(50),
    rorl INTEGER,
    rorx INTEGER,

    -- Configuration
    config_number BIGINT,
    attribute_number BIGINT,

    -- Project
    project_number VARCHAR(50),
    element_number VARCHAR(50),

    -- M3 Audit
    lmdt INTEGER,
    lmts BIGINT,

    -- Foreign Keys
    mo_id BIGINT REFERENCES manufacturing_orders(id) ON DELETE CASCADE,
    mop_id BIGINT REFERENCES planned_manufacturing_orders(id) ON DELETE CASCADE,

    -- Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT unique_production_order UNIQUE (order_number),
    CONSTRAINT check_order_type CHECK (order_type IN ('MO', 'MOP')),
    CONSTRAINT check_foreign_key CHECK (
        (order_type = 'MO' AND mo_id IS NOT NULL AND mop_id IS NULL) OR
        (order_type = 'MOP' AND mop_id IS NOT NULL AND mo_id IS NULL)
    )
);

-- Indexes for production_orders
CREATE INDEX idx_prod_facility ON production_orders(facility);
CREATE INDEX idx_prod_item ON production_orders(item_number);
CREATE INDEX idx_prod_dates ON production_orders(planned_start_date, planned_finish_date);
CREATE INDEX idx_prod_ref_order ON production_orders(rorc, rorn, rorl);
CREATE INDEX idx_prod_order_type ON production_orders(order_type);
CREATE INDEX idx_prod_warehouse ON production_orders(warehouse);
CREATE INDEX idx_prod_planner ON production_orders(planner_group);
