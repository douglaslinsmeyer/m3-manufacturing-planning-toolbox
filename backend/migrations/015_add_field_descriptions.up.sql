-- Migration 015: Add M3 field descriptions from Data Catalog metadata
-- These descriptions come from Infor Data Catalog and match M3 documentation

-- ============================================================================
-- Manufacturing Orders (MWOHED) Field Descriptions
-- ============================================================================

COMMENT ON COLUMN manufacturing_orders.cono IS 'Company';
COMMENT ON COLUMN manufacturing_orders.divi IS 'Division';
COMMENT ON COLUMN manufacturing_orders.faci IS 'Facility';
COMMENT ON COLUMN manufacturing_orders.mfno IS 'Manufacturing order number';
COMMENT ON COLUMN manufacturing_orders.prno IS 'Product';
COMMENT ON COLUMN manufacturing_orders.itno IS 'Item number';

COMMENT ON COLUMN manufacturing_orders.whst IS 'Status - manufacturing order';
COMMENT ON COLUMN manufacturing_orders.whhs IS 'Highest operation status - order';
COMMENT ON COLUMN manufacturing_orders.wmst IS 'Material status';
COMMENT ON COLUMN manufacturing_orders.mohs IS 'Hold MO status';

COMMENT ON COLUMN manufacturing_orders.orqt IS 'Ordered quantity - basic u/m';
COMMENT ON COLUMN manufacturing_orders.maqt IS 'Manufactured quantity';
COMMENT ON COLUMN manufacturing_orders.orqa IS 'Ordered quantity - alternate u/m';
COMMENT ON COLUMN manufacturing_orders.rvqt IS 'Received quantity';
COMMENT ON COLUMN manufacturing_orders.rvqa IS 'Received quantity (alternate u/m)';
COMMENT ON COLUMN manufacturing_orders.maqa IS 'Manufactured quantity (alternate u/m)';

COMMENT ON COLUMN manufacturing_orders.stdt IS 'Start date';
COMMENT ON COLUMN manufacturing_orders.fidt IS 'Finish date';
COMMENT ON COLUMN manufacturing_orders.msti IS 'Start time';
COMMENT ON COLUMN manufacturing_orders.mfti IS 'Finish time';
COMMENT ON COLUMN manufacturing_orders.fstd IS 'Original start date';
COMMENT ON COLUMN manufacturing_orders.ffid IS 'Original finish date';
COMMENT ON COLUMN manufacturing_orders.rsdt IS 'Actual start date';
COMMENT ON COLUMN manufacturing_orders.refd IS 'Actual finish date';
COMMENT ON COLUMN manufacturing_orders.rpdt IS 'Reporting date';

COMMENT ON COLUMN manufacturing_orders.prio IS 'Priority';
COMMENT ON COLUMN manufacturing_orders.resp IS 'Responsible';
COMMENT ON COLUMN manufacturing_orders.plgr IS 'Work center';
COMMENT ON COLUMN manufacturing_orders.wcln IS 'Production line';
COMMENT ON COLUMN manufacturing_orders.prdy IS 'Production days';

COMMENT ON COLUMN manufacturing_orders.whlo IS 'Warehouse';
COMMENT ON COLUMN manufacturing_orders.whsl IS 'Location';
COMMENT ON COLUMN manufacturing_orders.bano IS 'Lot number';

COMMENT ON COLUMN manufacturing_orders.rorc IS 'Reference order category';
COMMENT ON COLUMN manufacturing_orders.rorn IS 'Reference order number';
COMMENT ON COLUMN manufacturing_orders.rorl IS 'Reference order line';
COMMENT ON COLUMN manufacturing_orders.rorx IS 'Line suffix';

COMMENT ON COLUMN manufacturing_orders.prhl IS 'Product number highest level';
COMMENT ON COLUMN manufacturing_orders.mfhl IS 'MO number highest level';
COMMENT ON COLUMN manufacturing_orders.prlo IS 'Product number overlying level';
COMMENT ON COLUMN manufacturing_orders.mflo IS 'MO number next higher level';
COMMENT ON COLUMN manufacturing_orders.levl IS 'Lowest level';

COMMENT ON COLUMN manufacturing_orders.cfin IS 'Configuration number';
COMMENT ON COLUMN manufacturing_orders.atnr IS 'Attribute number';

COMMENT ON COLUMN manufacturing_orders.orty IS 'Order type';
COMMENT ON COLUMN manufacturing_orders.getp IS 'Origin';

COMMENT ON COLUMN manufacturing_orders.bdcd IS 'Explosion';
COMMENT ON COLUMN manufacturing_orders.scex IS 'Subcontracting exists';
COMMENT ON COLUMN manufacturing_orders.strt IS 'Product structure type';
COMMENT ON COLUMN manufacturing_orders.ecve IS 'Revision number';

COMMENT ON COLUMN manufacturing_orders.aoid IS 'Alternative routing';
COMMENT ON COLUMN manufacturing_orders.nuop IS 'Number of operations';
COMMENT ON COLUMN manufacturing_orders.nufo IS 'Number of finished operations';

COMMENT ON COLUMN manufacturing_orders.actp IS 'Action message (AM)';
COMMENT ON COLUMN manufacturing_orders.txt1 IS 'Text line 1';
COMMENT ON COLUMN manufacturing_orders.txt2 IS 'Text line 2';

COMMENT ON COLUMN manufacturing_orders.proj IS 'Project number';
COMMENT ON COLUMN manufacturing_orders.elno IS 'Project element';

COMMENT ON COLUMN manufacturing_orders.rgdt IS 'Entry date';
COMMENT ON COLUMN manufacturing_orders.rgtm IS 'Entry time';
COMMENT ON COLUMN manufacturing_orders.lmdt IS 'Change date';
COMMENT ON COLUMN manufacturing_orders.lmts IS 'Timestamp';
COMMENT ON COLUMN manufacturing_orders.chno IS 'Change number';
COMMENT ON COLUMN manufacturing_orders.chid IS 'Changed by';

COMMENT ON COLUMN manufacturing_orders.m3_timestamp IS 'Record modification time (Data Lake)';

COMMENT ON COLUMN manufacturing_orders.linked_co_number IS 'Demand order number (from MPREAL.DRDN)';
COMMENT ON COLUMN manufacturing_orders.linked_co_line IS 'Demand order line (from MPREAL.DRDL)';
COMMENT ON COLUMN manufacturing_orders.linked_co_suffix IS 'Line suffix (from MPREAL.DRDX)';
COMMENT ON COLUMN manufacturing_orders.allocated_qty IS 'Preallocated quantity (from MPREAL.PQTY)';

-- ============================================================================
-- Planned Manufacturing Orders (MMOPLP) Field Descriptions
-- ============================================================================

COMMENT ON COLUMN planned_manufacturing_orders.cono IS 'Company';
COMMENT ON COLUMN planned_manufacturing_orders.divi IS 'Division';
COMMENT ON COLUMN planned_manufacturing_orders.faci IS 'Facility';
COMMENT ON COLUMN planned_manufacturing_orders.plpn IS 'Planned order';
COMMENT ON COLUMN planned_manufacturing_orders.plps IS 'Subnumber - planned order';
COMMENT ON COLUMN planned_manufacturing_orders.prno IS 'Product';
COMMENT ON COLUMN planned_manufacturing_orders.itno IS 'Item number';

COMMENT ON COLUMN planned_manufacturing_orders.psts IS 'Status - planned MO';
COMMENT ON COLUMN planned_manufacturing_orders.whst IS 'Status - manufacturing order';
COMMENT ON COLUMN planned_manufacturing_orders.actp IS 'Action message (AM)';

COMMENT ON COLUMN planned_manufacturing_orders.orty IS 'Order type';
COMMENT ON COLUMN planned_manufacturing_orders.gety IS 'Generation reference';

COMMENT ON COLUMN planned_manufacturing_orders.ppqt IS 'Planned quantity';
COMMENT ON COLUMN planned_manufacturing_orders.orqa IS 'Ordered quantity - alternate u/m';

COMMENT ON COLUMN planned_manufacturing_orders.reld IS 'Release date';
COMMENT ON COLUMN planned_manufacturing_orders.stdt IS 'Start date';
COMMENT ON COLUMN planned_manufacturing_orders.fidt IS 'Finish date';
COMMENT ON COLUMN planned_manufacturing_orders.msti IS 'Start time';
COMMENT ON COLUMN planned_manufacturing_orders.mfti IS 'Finish time';
COMMENT ON COLUMN planned_manufacturing_orders.pldt IS 'Planning date';

COMMENT ON COLUMN planned_manufacturing_orders.resp IS 'Responsible';
COMMENT ON COLUMN planned_manufacturing_orders.prip IS 'Priority';
COMMENT ON COLUMN planned_manufacturing_orders.plgr IS 'Work center';
COMMENT ON COLUMN planned_manufacturing_orders.wcln IS 'Production line';
COMMENT ON COLUMN planned_manufacturing_orders.prdy IS 'Production days';

COMMENT ON COLUMN planned_manufacturing_orders.whlo IS 'Warehouse';

COMMENT ON COLUMN planned_manufacturing_orders.rorc IS 'Reference order category';
COMMENT ON COLUMN planned_manufacturing_orders.rorn IS 'Reference order number';
COMMENT ON COLUMN planned_manufacturing_orders.rorl IS 'Reference order line';
COMMENT ON COLUMN planned_manufacturing_orders.rorx IS 'Line suffix';
COMMENT ON COLUMN planned_manufacturing_orders.rorh IS 'Reference order number (header)';

COMMENT ON COLUMN planned_manufacturing_orders.pllo IS 'Proposal number - overlying level';
COMMENT ON COLUMN planned_manufacturing_orders.plhl IS 'Proposal number - highest level';

COMMENT ON COLUMN planned_manufacturing_orders.atnr IS 'Attribute number';
COMMENT ON COLUMN planned_manufacturing_orders.cfin IS 'Configuration number';

COMMENT ON COLUMN planned_manufacturing_orders.proj IS 'Project number';
COMMENT ON COLUMN planned_manufacturing_orders.elno IS 'Project element';

COMMENT ON COLUMN planned_manufacturing_orders.messages IS 'Warning messages (MSG1-MSG4) as JSONB';

COMMENT ON COLUMN planned_manufacturing_orders.nuau IS 'Number of automatic updates';
COMMENT ON COLUMN planned_manufacturing_orders.ordp IS 'Order dependent';

COMMENT ON COLUMN planned_manufacturing_orders.rgdt IS 'Entry date';
COMMENT ON COLUMN planned_manufacturing_orders.rgtm IS 'Entry time';
COMMENT ON COLUMN planned_manufacturing_orders.lmdt IS 'Change date';
COMMENT ON COLUMN planned_manufacturing_orders.lmts IS 'Timestamp';
COMMENT ON COLUMN planned_manufacturing_orders.chno IS 'Change number';
COMMENT ON COLUMN planned_manufacturing_orders.chid IS 'Changed by';

COMMENT ON COLUMN planned_manufacturing_orders.m3_timestamp IS 'Record modification time (Data Lake)';

COMMENT ON COLUMN planned_manufacturing_orders.linked_co_number IS 'Demand order number (from MPREAL.DRDN)';
COMMENT ON COLUMN planned_manufacturing_orders.linked_co_line IS 'Demand order line (from MPREAL.DRDL)';
COMMENT ON COLUMN planned_manufacturing_orders.linked_co_suffix IS 'Line suffix (from MPREAL.DRDX)';
COMMENT ON COLUMN planned_manufacturing_orders.allocated_qty IS 'Preallocated quantity (from MPREAL.PQTY)';

-- ============================================================================
-- Production Orders Field Descriptions
-- ============================================================================

COMMENT ON COLUMN production_orders.order_type IS 'Discriminator: MO or MOP';
COMMENT ON COLUMN production_orders.order_number IS 'Manufacturing order or planned order number';

COMMENT ON COLUMN production_orders.cono IS 'Company';
COMMENT ON COLUMN production_orders.divi IS 'Division';
COMMENT ON COLUMN production_orders.faci IS 'Facility';
COMMENT ON COLUMN production_orders.prno IS 'Product';
COMMENT ON COLUMN production_orders.itno IS 'Item number';

COMMENT ON COLUMN production_orders.ordered_quantity IS 'Ordered/planned quantity';
COMMENT ON COLUMN production_orders.manufactured_quantity IS 'Manufactured quantity (MO only)';

COMMENT ON COLUMN production_orders.planned_start_date IS 'Start date (YYYYMMDD)';
COMMENT ON COLUMN production_orders.planned_finish_date IS 'Finish date (YYYYMMDD)';
COMMENT ON COLUMN production_orders.actual_start_date IS 'Actual start date (MO only)';
COMMENT ON COLUMN production_orders.actual_finish_date IS 'Actual finish date (MO only)';
COMMENT ON COLUMN production_orders.release_date IS 'Release date (MOP only)';
COMMENT ON COLUMN production_orders.material_start_date IS 'Material start time';
COMMENT ON COLUMN production_orders.material_finish_date IS 'Material finish time';

COMMENT ON COLUMN production_orders.status IS 'MO/MOP status';
COMMENT ON COLUMN production_orders.proposal_status IS 'Planned order status (MOP only)';

COMMENT ON COLUMN production_orders.priority IS 'Priority';
COMMENT ON COLUMN production_orders.responsible IS 'Responsible';
COMMENT ON COLUMN production_orders.planner_group IS 'Work center';
COMMENT ON COLUMN production_orders.production_line IS 'Production line';

COMMENT ON COLUMN production_orders.warehouse IS 'Warehouse';
COMMENT ON COLUMN production_orders.location IS 'Location (MO only)';
COMMENT ON COLUMN production_orders.batch_number IS 'Lot number (MO only)';

COMMENT ON COLUMN production_orders.rorc IS 'Reference order category';
COMMENT ON COLUMN production_orders.rorn IS 'Reference order number';
COMMENT ON COLUMN production_orders.rorl IS 'Reference order line';
COMMENT ON COLUMN production_orders.rorx IS 'Line suffix';

COMMENT ON COLUMN production_orders.config_number IS 'Configuration number';
COMMENT ON COLUMN production_orders.attribute_number IS 'Attribute number';

COMMENT ON COLUMN production_orders.project_number IS 'Project number';
COMMENT ON COLUMN production_orders.element_number IS 'Project element';

COMMENT ON COLUMN production_orders.lmdt IS 'Change date';
COMMENT ON COLUMN production_orders.lmts IS 'Timestamp';

COMMENT ON COLUMN production_orders.mo_id IS 'Foreign key to manufacturing_orders.id';
COMMENT ON COLUMN production_orders.mop_id IS 'Foreign key to planned_manufacturing_orders.id';
