-- Migration 016: Rebuild customer_order_lines with M3 naming and essential fields
-- All M3 fields stored as VARCHAR to match Data Fabric source format
-- ~110 essential fields for manufacturing planning and order management

DROP TABLE IF EXISTS customer_order_lines CASCADE;

CREATE TABLE customer_order_lines (
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to customer_orders (application-level)
    co_id BIGINT REFERENCES customer_orders(id) ON DELETE CASCADE,

    -- M3 Core Identifiers
    cono VARCHAR(10) NOT NULL,
    divi VARCHAR(10),
    orno VARCHAR(50) NOT NULL,
    ponr VARCHAR(10) NOT NULL,
    posx VARCHAR(10) NOT NULL,

    -- M3 Item Information (ALL)
    itno VARCHAR(50) NOT NULL,
    itds TEXT,
    teds TEXT,
    repi VARCHAR(50),

    -- M3 Status/Type
    orst VARCHAR(10),
    orty VARCHAR(10),

    -- M3 Facility/Warehouse
    faci VARCHAR(10),
    whlo VARCHAR(10),

    -- M3 Quantities - Basic U/M
    orqt VARCHAR(30),
    rnqt VARCHAR(30),
    alqt VARCHAR(30),
    dlqt VARCHAR(30),
    ivqt VARCHAR(30),

    -- M3 Quantities - Alternate U/M
    orqa VARCHAR(30),
    rnqa VARCHAR(30),
    alqa VARCHAR(30),
    dlqa VARCHAR(30),
    ivqa VARCHAR(30),

    -- M3 Units
    alun VARCHAR(10),
    cofa VARCHAR(30),
    spun VARCHAR(10),

    -- M3 Delivery Dates
    dwdt VARCHAR(10),
    dwhm VARCHAR(10),
    codt VARCHAR(10),
    cohm VARCHAR(10),
    pldt VARCHAR(10),
    fded VARCHAR(10),
    lded VARCHAR(10),

    -- M3 Pricing
    sapr VARCHAR(30),
    nepr VARCHAR(30),
    lnam VARCHAR(30),
    cucd VARCHAR(10),

    -- M3 Discounts (DIP1-DIP6, DIA1-DIA6 - all as strings with decimals)
    dip1 VARCHAR(30),
    dip2 VARCHAR(30),
    dip3 VARCHAR(30),
    dip4 VARCHAR(30),
    dip5 VARCHAR(30),
    dip6 VARCHAR(30),
    dia1 VARCHAR(30),
    dia2 VARCHAR(30),
    dia3 VARCHAR(30),
    dia4 VARCHAR(30),
    dia5 VARCHAR(30),
    dia6 VARCHAR(30),

    -- M3 Reference Orders
    rorc VARCHAR(10),
    rorn VARCHAR(50),
    rorl VARCHAR(10),
    rorx VARCHAR(10),

    -- M3 Customer References (ALL)
    cuno VARCHAR(50),
    cuor VARCHAR(50),
    cupo VARCHAR(10),
    cusx VARCHAR(10),

    -- M3 Product/Model (ALL)
    prno VARCHAR(50),
    hdpr VARCHAR(50),
    popn VARCHAR(50),
    alwt VARCHAR(10),
    alwq VARCHAR(50),

    -- M3 Delivery/Route (ALL)
    adid VARCHAR(50),
    rout VARCHAR(50),
    rodn VARCHAR(10),
    dsdt VARCHAR(10),
    dshm VARCHAR(10),
    modl VARCHAR(10),
    tedl VARCHAR(50),
    tel2 TEXT,

    -- M3 Packaging (ALL)
    tepa VARCHAR(50),
    pact VARCHAR(50),
    cupa VARCHAR(50),

    -- M3 Partner/EDI (ALL)
    e0pa VARCHAR(50),
    dsgp VARCHAR(50),
    pusn VARCHAR(50),
    putp VARCHAR(10),

    -- M3 Attributes (ATV1-ATV0)
    atv1 VARCHAR(30),
    atv2 VARCHAR(30),
    atv3 VARCHAR(30),
    atv4 VARCHAR(30),
    atv5 VARCHAR(30),
    atv6 VARCHAR(50),
    atv7 VARCHAR(50),
    atv8 VARCHAR(50),
    atv9 VARCHAR(50),
    atv0 VARCHAR(50),

    -- M3 User-Defined Alpha (UCA1-UCA0)
    uca1 VARCHAR(50),
    uca2 VARCHAR(50),
    uca3 VARCHAR(50),
    uca4 VARCHAR(50),
    uca5 VARCHAR(50),
    uca6 VARCHAR(50),
    uca7 VARCHAR(50),
    uca8 VARCHAR(50),
    uca9 VARCHAR(50),
    uca0 VARCHAR(50),

    -- M3 User-Defined Numeric (UDN1-UDN6)
    udn1 VARCHAR(30),
    udn2 VARCHAR(30),
    udn3 VARCHAR(30),
    udn4 VARCHAR(30),
    udn5 VARCHAR(30),
    udn6 VARCHAR(30),

    -- M3 User-Defined Date (UID1-UID3)
    uid1 VARCHAR(10),
    uid2 VARCHAR(10),
    uid3 VARCHAR(10),

    -- M3 User-Defined Text
    uct1 TEXT,

    -- M3 Configuration/Attributes
    atnr VARCHAR(20),
    atmo VARCHAR(20),
    atpr VARCHAR(20),
    cfin VARCHAR(20),

    -- M3 Project
    proj VARCHAR(50),
    elno VARCHAR(50),

    -- M3 Audit Fields
    rgdt VARCHAR(10),
    rgtm VARCHAR(10),
    lmdt VARCHAR(10),
    chno VARCHAR(10),
    chid VARCHAR(50),
    lmts VARCHAR(30),

    -- Data Lake Metadata
    m3_timestamp TEXT,

    -- Application Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT unique_orno_ponr_posx UNIQUE (orno, ponr, posx)
);

-- Indexes
CREATE INDEX idx_co_lines_cono_divi ON customer_order_lines(cono, divi);
CREATE INDEX idx_co_lines_orno ON customer_order_lines(orno);
CREATE INDEX idx_co_lines_itno ON customer_order_lines(itno);
CREATE INDEX idx_co_lines_orst ON customer_order_lines(orst);
CREATE INDEX idx_co_lines_faci ON customer_order_lines(faci);
CREATE INDEX idx_co_lines_dates ON customer_order_lines(dwdt, codt);
CREATE INDEX idx_co_lines_lmdt ON customer_order_lines(lmdt);
CREATE INDEX idx_co_lines_rorc_rorn ON customer_order_lines(rorc, rorn, rorl, rorx);
CREATE INDEX idx_co_lines_cuno ON customer_order_lines(cuno);
CREATE INDEX idx_co_lines_co_id ON customer_order_lines(co_id);

-- Update trigger
CREATE TRIGGER update_customer_order_lines_updated_at
    BEFORE UPDATE ON customer_order_lines
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Field Descriptions from M3 Data Catalog (OOLINE table)
-- ============================================================================

-- Core Identifiers
COMMENT ON COLUMN customer_order_lines.cono IS 'Company';
COMMENT ON COLUMN customer_order_lines.divi IS 'Division';
COMMENT ON COLUMN customer_order_lines.orno IS 'Customer order number';
COMMENT ON COLUMN customer_order_lines.ponr IS 'Order line number';
COMMENT ON COLUMN customer_order_lines.posx IS 'Line suffix';

-- Item Information
COMMENT ON COLUMN customer_order_lines.itno IS 'Item number';
COMMENT ON COLUMN customer_order_lines.itds IS 'Name';
COMMENT ON COLUMN customer_order_lines.teds IS 'Description 1';
COMMENT ON COLUMN customer_order_lines.repi IS 'Replaced item number';

-- Status/Type
COMMENT ON COLUMN customer_order_lines.orst IS 'Highest status - customer order';
COMMENT ON COLUMN customer_order_lines.orty IS 'Order type';

-- Facility/Warehouse
COMMENT ON COLUMN customer_order_lines.faci IS 'Facility';
COMMENT ON COLUMN customer_order_lines.whlo IS 'Warehouse';

-- Quantities - Basic U/M
COMMENT ON COLUMN customer_order_lines.orqt IS 'Ordered quantity - basic u/m';
COMMENT ON COLUMN customer_order_lines.rnqt IS 'Remaining quantity - basic u/m';
COMMENT ON COLUMN customer_order_lines.alqt IS 'Allocated quantity - basic u/m';
COMMENT ON COLUMN customer_order_lines.dlqt IS 'Delivered quantity - basic u/m';
COMMENT ON COLUMN customer_order_lines.ivqt IS 'Invoiced quantity - basic u/m';

-- Quantities - Alternate U/M
COMMENT ON COLUMN customer_order_lines.orqa IS 'Ordered quantity - alternate u/m';
COMMENT ON COLUMN customer_order_lines.rnqa IS 'Remaining quantity - alternate u/m';
COMMENT ON COLUMN customer_order_lines.alqa IS 'Allocated quantity - alternate u/m';
COMMENT ON COLUMN customer_order_lines.dlqa IS 'Delivered quantity - alternate u/m';
COMMENT ON COLUMN customer_order_lines.ivqa IS 'Invoiced quantity - alternate u/m';

-- Units
COMMENT ON COLUMN customer_order_lines.alun IS 'Alternate u/m';
COMMENT ON COLUMN customer_order_lines.cofa IS 'Conversion factor';
COMMENT ON COLUMN customer_order_lines.spun IS 'Sales price unit of measure';

-- Delivery Dates
COMMENT ON COLUMN customer_order_lines.dwdt IS 'Requested delivery date';
COMMENT ON COLUMN customer_order_lines.dwhm IS 'Requested delivery time';
COMMENT ON COLUMN customer_order_lines.codt IS 'Confirmed delivery date';
COMMENT ON COLUMN customer_order_lines.cohm IS 'Confirmed delivery time';
COMMENT ON COLUMN customer_order_lines.pldt IS 'Planning date';
COMMENT ON COLUMN customer_order_lines.fded IS 'First delivery date';
COMMENT ON COLUMN customer_order_lines.lded IS 'Last delivery';

-- Pricing
COMMENT ON COLUMN customer_order_lines.sapr IS 'Sales price';
COMMENT ON COLUMN customer_order_lines.nepr IS 'Net price';
COMMENT ON COLUMN customer_order_lines.lnam IS 'Line amount - order currency';
COMMENT ON COLUMN customer_order_lines.cucd IS 'Currency';

-- Discounts
COMMENT ON COLUMN customer_order_lines.dip1 IS 'Discount 1';
COMMENT ON COLUMN customer_order_lines.dip2 IS 'Discount 2';
COMMENT ON COLUMN customer_order_lines.dip3 IS 'Discount 3';
COMMENT ON COLUMN customer_order_lines.dip4 IS 'Discount 4';
COMMENT ON COLUMN customer_order_lines.dip5 IS 'Discount 5';
COMMENT ON COLUMN customer_order_lines.dip6 IS 'Discount 6';
COMMENT ON COLUMN customer_order_lines.dia1 IS 'Discount amount 1';
COMMENT ON COLUMN customer_order_lines.dia2 IS 'Discount amount 2';
COMMENT ON COLUMN customer_order_lines.dia3 IS 'Discount amount 3';
COMMENT ON COLUMN customer_order_lines.dia4 IS 'Discount amount 4';
COMMENT ON COLUMN customer_order_lines.dia5 IS 'Discount amount 5';
COMMENT ON COLUMN customer_order_lines.dia6 IS 'Discount amount 6';

-- Reference Orders
COMMENT ON COLUMN customer_order_lines.rorc IS 'Reference order category';
COMMENT ON COLUMN customer_order_lines.rorn IS 'Reference order number';
COMMENT ON COLUMN customer_order_lines.rorl IS 'Reference order line';
COMMENT ON COLUMN customer_order_lines.rorx IS 'Line suffix';

-- Customer References
COMMENT ON COLUMN customer_order_lines.cuno IS 'Customer';
COMMENT ON COLUMN customer_order_lines.cuor IS 'Customer''s order number';
COMMENT ON COLUMN customer_order_lines.cupo IS 'Customer order line number';
COMMENT ON COLUMN customer_order_lines.cusx IS 'Customer line suffix';

-- Product/Model
COMMENT ON COLUMN customer_order_lines.prno IS 'Product';
COMMENT ON COLUMN customer_order_lines.hdpr IS 'Main product';
COMMENT ON COLUMN customer_order_lines.popn IS 'Alias number';
COMMENT ON COLUMN customer_order_lines.alwt IS 'Alias category';
COMMENT ON COLUMN customer_order_lines.alwq IS 'Alias qualifier';

-- Delivery/Route
COMMENT ON COLUMN customer_order_lines.adid IS 'Address number';
COMMENT ON COLUMN customer_order_lines.rout IS 'Route';
COMMENT ON COLUMN customer_order_lines.rodn IS 'Route departure';
COMMENT ON COLUMN customer_order_lines.dsdt IS 'Departure date';
COMMENT ON COLUMN customer_order_lines.dshm IS 'Departure time';
COMMENT ON COLUMN customer_order_lines.modl IS 'Delivery method';
COMMENT ON COLUMN customer_order_lines.tedl IS 'Delivery terms';
COMMENT ON COLUMN customer_order_lines.tel2 IS 'Terms text';

-- Packaging
COMMENT ON COLUMN customer_order_lines.tepa IS 'Packaging terms';
COMMENT ON COLUMN customer_order_lines.pact IS 'Packaging';
COMMENT ON COLUMN customer_order_lines.cupa IS 'Customer''s packaging identity';

-- Partner/EDI
COMMENT ON COLUMN customer_order_lines.e0pa IS 'Partner';
COMMENT ON COLUMN customer_order_lines.dsgp IS 'Delivery schedule group';
COMMENT ON COLUMN customer_order_lines.pusn IS 'Delivery note reference';
COMMENT ON COLUMN customer_order_lines.putp IS 'Delivery note reference qualifier';

-- Attributes (ATV1-ATV0)
COMMENT ON COLUMN customer_order_lines.atv1 IS 'Attribute value - display field 1';
COMMENT ON COLUMN customer_order_lines.atv2 IS 'Attribute value - display field 2';
COMMENT ON COLUMN customer_order_lines.atv3 IS 'Attribute value - display field 3';
COMMENT ON COLUMN customer_order_lines.atv4 IS 'Attribute value - display field 4';
COMMENT ON COLUMN customer_order_lines.atv5 IS 'Attribute value - display field 5';
COMMENT ON COLUMN customer_order_lines.atv6 IS 'Attribute value - display field 6';
COMMENT ON COLUMN customer_order_lines.atv7 IS 'Attribute value - display field 7';
COMMENT ON COLUMN customer_order_lines.atv8 IS 'Attribute value - display field 8';
COMMENT ON COLUMN customer_order_lines.atv9 IS 'Attribute value - display field 9';
COMMENT ON COLUMN customer_order_lines.atv0 IS 'Attribute value - display field 10';

-- User-Defined Alpha Fields
COMMENT ON COLUMN customer_order_lines.uca1 IS 'User-defined alpha field 1';
COMMENT ON COLUMN customer_order_lines.uca2 IS 'User-defined alpha field 2';
COMMENT ON COLUMN customer_order_lines.uca3 IS 'User-defined alpha field 3';
COMMENT ON COLUMN customer_order_lines.uca4 IS 'User-defined alpha field 4';
COMMENT ON COLUMN customer_order_lines.uca5 IS 'User-defined alpha field 5';
COMMENT ON COLUMN customer_order_lines.uca6 IS 'User-defined alpha field 6';
COMMENT ON COLUMN customer_order_lines.uca7 IS 'User-defined alpha field 7';
COMMENT ON COLUMN customer_order_lines.uca8 IS 'User-defined alpha field 8';
COMMENT ON COLUMN customer_order_lines.uca9 IS 'User-defined alpha field 9';
COMMENT ON COLUMN customer_order_lines.uca0 IS 'User-defined alpha field 10';

-- User-Defined Numeric Fields
COMMENT ON COLUMN customer_order_lines.udn1 IS 'User-defined numeric 1';
COMMENT ON COLUMN customer_order_lines.udn2 IS 'User-defined numeric 2';
COMMENT ON COLUMN customer_order_lines.udn3 IS 'User-defined numeric 3';
COMMENT ON COLUMN customer_order_lines.udn4 IS 'User-defined numeric 4';
COMMENT ON COLUMN customer_order_lines.udn5 IS 'User-defined numeric 5';
COMMENT ON COLUMN customer_order_lines.udn6 IS 'User-defined numeric 6';

-- User-Defined Date Fields
COMMENT ON COLUMN customer_order_lines.uid1 IS 'User-defined date 1';
COMMENT ON COLUMN customer_order_lines.uid2 IS 'User-defined date 2';
COMMENT ON COLUMN customer_order_lines.uid3 IS 'User-defined date 3';

-- User-Defined Text
COMMENT ON COLUMN customer_order_lines.uct1 IS 'User-defined text field 1';

-- Configuration/Attributes
COMMENT ON COLUMN customer_order_lines.atnr IS 'Attribute number';
COMMENT ON COLUMN customer_order_lines.atmo IS 'Attribute model';
COMMENT ON COLUMN customer_order_lines.atpr IS 'Attribute pricing rule';
COMMENT ON COLUMN customer_order_lines.cfin IS 'Configuration number';

-- Project
COMMENT ON COLUMN customer_order_lines.proj IS 'Project number';
COMMENT ON COLUMN customer_order_lines.elno IS 'Project element';

-- M3 Audit Fields
COMMENT ON COLUMN customer_order_lines.rgdt IS 'Entry date';
COMMENT ON COLUMN customer_order_lines.rgtm IS 'Entry time';
COMMENT ON COLUMN customer_order_lines.lmdt IS 'Change date';
COMMENT ON COLUMN customer_order_lines.chno IS 'Change number';
COMMENT ON COLUMN customer_order_lines.chid IS 'Changed by';
COMMENT ON COLUMN customer_order_lines.lmts IS 'Timestamp';

-- Data Lake Metadata
COMMENT ON COLUMN customer_order_lines.m3_timestamp IS 'Record modification time (Data Lake)';

-- Application Metadata
COMMENT ON COLUMN customer_order_lines.co_id IS 'Foreign key to customer_orders table';
