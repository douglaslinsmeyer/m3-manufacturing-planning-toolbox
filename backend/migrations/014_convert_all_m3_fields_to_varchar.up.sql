-- Migration 014: Convert ALL M3 fields to VARCHAR to match Data Fabric source format
-- Data Fabric returns nearly all fields as strings, so store them as received
-- Only keep typed: id (BIGSERIAL), timestamps (TIMESTAMP), is_deleted (BOOLEAN)

-- Drop and rebuild with string-based schema
DROP TABLE IF EXISTS production_orders CASCADE;
DROP TABLE IF EXISTS manufacturing_orders CASCADE;
DROP TABLE IF EXISTS planned_manufacturing_orders CASCADE;

-- ============================================================================
-- Manufacturing Orders - All M3 fields as VARCHAR
-- ============================================================================

CREATE TABLE manufacturing_orders (
    id BIGSERIAL PRIMARY KEY,

    -- M3 Core Identifiers
    cono VARCHAR(10) NOT NULL,
    divi VARCHAR(10),
    faci VARCHAR(10) NOT NULL,
    mfno VARCHAR(50) NOT NULL,
    prno VARCHAR(50),
    itno VARCHAR(50) NOT NULL,

    -- M3 Status Fields
    whst VARCHAR(10),
    whhs VARCHAR(10),
    wmst VARCHAR(10),
    mohs VARCHAR(10),

    -- M3 Quantities (as strings from Data Fabric)
    orqt VARCHAR(30),
    maqt VARCHAR(30),
    orqa VARCHAR(30),
    rvqt VARCHAR(30),
    rvqa VARCHAR(30),
    maqa VARCHAR(30),

    -- M3 Date Fields (as strings YYYYMMDD)
    stdt VARCHAR(10),
    fidt VARCHAR(10),
    msti VARCHAR(10),
    mfti VARCHAR(10),
    fstd VARCHAR(10),
    ffid VARCHAR(10),
    rsdt VARCHAR(10),
    refd VARCHAR(10),
    rpdt VARCHAR(10),

    -- M3 Planning Fields
    prio VARCHAR(10),
    resp VARCHAR(50),
    plgr VARCHAR(50),
    wcln VARCHAR(50),
    prdy VARCHAR(10),

    -- M3 Warehouse/Location
    whlo VARCHAR(10),
    whsl VARCHAR(50),
    bano VARCHAR(50),

    -- M3 Reference Orders
    rorc VARCHAR(10),
    rorn VARCHAR(50),
    rorl VARCHAR(10),
    rorx VARCHAR(10),

    -- M3 Hierarchy
    prhl VARCHAR(50),
    mfhl VARCHAR(50),
    prlo VARCHAR(50),
    mflo VARCHAR(50),
    levl VARCHAR(10),

    -- M3 Configuration/Attributes
    cfin VARCHAR(20),
    atnr VARCHAR(20),

    -- M3 Order Type/Origin
    orty VARCHAR(10),
    getp VARCHAR(10),

    -- M3 Material/BOM
    bdcd VARCHAR(10),
    scex VARCHAR(10),
    strt VARCHAR(10),
    ecve VARCHAR(50),

    -- M3 Routing
    aoid VARCHAR(50),
    nuop VARCHAR(10),
    nufo VARCHAR(10),

    -- M3 Action/Text
    actp VARCHAR(10),
    txt1 TEXT,
    txt2 TEXT,

    -- M3 Project
    proj VARCHAR(50),
    elno VARCHAR(50),

    -- M3 Audit Fields
    rgdt VARCHAR(10),
    rgtm VARCHAR(10),
    lmdt VARCHAR(10),
    lmts VARCHAR(30),
    chno VARCHAR(10),
    chid VARCHAR(50),

    -- Data Lake Metadata (as string from Data Fabric)
    m3_timestamp TEXT,

    -- CO Link (from MPREAL join - all as strings)
    linked_co_number VARCHAR(50),
    linked_co_line VARCHAR(10),
    linked_co_suffix VARCHAR(10),
    allocated_qty VARCHAR(30),

    -- Application Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT unique_mfno UNIQUE (mfno)
);

CREATE INDEX idx_mo_faci ON manufacturing_orders(faci);
CREATE INDEX idx_mo_itno ON manufacturing_orders(itno);
CREATE INDEX idx_mo_status ON manufacturing_orders(whst);
CREATE INDEX idx_mo_dates ON manufacturing_orders(stdt, fidt);
CREATE INDEX idx_mo_lmdt ON manufacturing_orders(lmdt);
CREATE INDEX idx_mo_ref_order ON manufacturing_orders(rorc, rorn, rorl);
CREATE INDEX idx_mo_warehouse ON manufacturing_orders(whlo);
CREATE INDEX idx_mo_planner ON manufacturing_orders(plgr);
CREATE INDEX idx_mo_linked_co ON manufacturing_orders(linked_co_number, linked_co_line);

-- ============================================================================
-- Planned Manufacturing Orders - All M3 fields as VARCHAR
-- ============================================================================

CREATE TABLE planned_manufacturing_orders (
    id BIGSERIAL PRIMARY KEY,

    -- M3 Core Identifiers
    cono VARCHAR(10) NOT NULL,
    divi VARCHAR(10),
    faci VARCHAR(10) NOT NULL,
    plpn VARCHAR(20) NOT NULL,
    plps VARCHAR(10),
    prno VARCHAR(50),
    itno VARCHAR(50) NOT NULL,

    -- M3 Status Fields
    psts VARCHAR(10),
    whst VARCHAR(10),
    actp VARCHAR(10),

    -- M3 Order Type
    orty VARCHAR(10),
    gety VARCHAR(10),

    -- M3 Quantities (as strings)
    ppqt VARCHAR(30),
    orqa VARCHAR(30),

    -- M3 Date Fields (as strings YYYYMMDD)
    reld VARCHAR(10),
    stdt VARCHAR(10),
    fidt VARCHAR(10),
    msti VARCHAR(10),
    mfti VARCHAR(10),
    pldt VARCHAR(10),

    -- M3 Planning Fields
    resp VARCHAR(50),
    prip VARCHAR(10),
    plgr VARCHAR(50),
    wcln VARCHAR(50),
    prdy VARCHAR(10),

    -- M3 Warehouse
    whlo VARCHAR(10),

    -- M3 Reference Orders
    rorc VARCHAR(10),
    rorn VARCHAR(50),
    rorl VARCHAR(10),
    rorx VARCHAR(10),
    rorh VARCHAR(50),

    -- M3 Hierarchy
    pllo VARCHAR(50),
    plhl VARCHAR(50),

    -- M3 Configuration/Attributes
    atnr VARCHAR(20),
    cfin VARCHAR(20),

    -- M3 Project
    proj VARCHAR(50),
    elno VARCHAR(50),

    -- M3 Messages (JSONB)
    messages JSONB,

    -- M3 Planning Parameters
    nuau VARCHAR(10),
    ordp VARCHAR(10),

    -- M3 Audit Fields
    rgdt VARCHAR(10),
    rgtm VARCHAR(10),
    lmdt VARCHAR(10),
    lmts VARCHAR(30),
    chno VARCHAR(10),
    chid VARCHAR(50),

    -- Data Lake Metadata
    m3_timestamp TEXT,

    -- CO Link (from MPREAL - all as strings)
    linked_co_number VARCHAR(50),
    linked_co_line VARCHAR(10),
    linked_co_suffix VARCHAR(10),
    allocated_qty VARCHAR(30),

    -- Application Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT unique_plpn UNIQUE (plpn)
);

CREATE INDEX idx_mop_faci ON planned_manufacturing_orders(faci);
CREATE INDEX idx_mop_itno ON planned_manufacturing_orders(itno);
CREATE INDEX idx_mop_status ON planned_manufacturing_orders(psts);
CREATE INDEX idx_mop_dates ON planned_manufacturing_orders(stdt, fidt);
CREATE INDEX idx_mop_lmdt ON planned_manufacturing_orders(lmdt);
CREATE INDEX idx_mop_linked_co ON planned_manufacturing_orders(linked_co_number, linked_co_line);
CREATE INDEX idx_mop_warehouse ON planned_manufacturing_orders(whlo);
CREATE INDEX idx_mop_planner ON planned_manufacturing_orders(plgr);

-- ============================================================================
-- Production Orders - Unified View with string-based M3 fields
-- ============================================================================

CREATE TABLE production_orders (
    id BIGSERIAL PRIMARY KEY,

    -- Discriminator
    order_type VARCHAR(10) NOT NULL,
    order_number VARCHAR(50) NOT NULL,

    -- M3 Context
    cono VARCHAR(10) NOT NULL,
    divi VARCHAR(10),
    faci VARCHAR(10) NOT NULL,

    -- Item Information
    prno VARCHAR(50),
    itno VARCHAR(50) NOT NULL,

    -- Quantities (as strings)
    ordered_quantity VARCHAR(30),
    manufactured_quantity VARCHAR(30),

    -- Dates (as strings YYYYMMDD - use CAST to convert)
    planned_start_date VARCHAR(10),
    planned_finish_date VARCHAR(10),
    actual_start_date VARCHAR(10),
    actual_finish_date VARCHAR(10),
    release_date VARCHAR(10),
    material_start_date VARCHAR(10),
    material_finish_date VARCHAR(10),

    -- Status
    status VARCHAR(10),
    proposal_status VARCHAR(10),

    -- Planning
    priority VARCHAR(10),
    responsible VARCHAR(50),
    planner_group VARCHAR(50),
    production_line VARCHAR(50),

    -- Warehouse/Location
    warehouse VARCHAR(10),
    location VARCHAR(50),
    batch_number VARCHAR(50),

    -- Reference Orders
    rorc VARCHAR(10),
    rorn VARCHAR(50),
    rorl VARCHAR(10),
    rorx VARCHAR(10),

    -- Configuration
    config_number VARCHAR(20),
    attribute_number VARCHAR(20),

    -- Project
    project_number VARCHAR(50),
    element_number VARCHAR(50),

    -- M3 Audit
    lmdt VARCHAR(10),
    lmts VARCHAR(30),

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

CREATE INDEX idx_prod_faci ON production_orders(faci);
CREATE INDEX idx_prod_itno ON production_orders(itno);
CREATE INDEX idx_prod_dates ON production_orders(planned_start_date, planned_finish_date);
CREATE INDEX idx_prod_ref_order ON production_orders(rorc, rorn, rorl);
CREATE INDEX idx_prod_order_type ON production_orders(order_type);
CREATE INDEX idx_prod_warehouse ON production_orders(warehouse);
CREATE INDEX idx_prod_planner ON production_orders(planner_group);
