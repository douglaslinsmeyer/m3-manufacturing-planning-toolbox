# Planned Manufacturing Order (MMOPLP) - Schema Map

## Table Overview
**M3 Table**: MMOPLP
**Description**: Planned Manufacturing Order (MOP) - MRP-generated manufacturing proposals
**Record Count**: 91 fields

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| FACI | string | Facility | Manufacturing facility code |
| PLPN | integer | Planned Order Number | MOP number (unique, max: 9999999) |
| PLPS | integer | Subnumber | Sub-order number for splits (max: 999) |
| PRNO | string | Product Number | Product to be manufactured |
| ITNO | string | Item Number | Item number (same as product) |
| WHLO | string | Warehouse | Warehouse for finished goods |

**Note**: PLPN + PLPS forms unique identifier. PLPS is used when MRP splits orders.

---

## Status Fields

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PSTS | string | Proposal Status | Status of the planned order |
| WHST | string | MO Status | Status if converted to MO |

**Common PSTS (Proposal Status) Values:**
- 10 = New Proposal
- 20 = Acknowledged
- 30 = Converted to MO
- 90 = Deleted/Cancelled

---

## Action Messages

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ACTP | string | Action Message | MRP action message code |

**Common ACTP Values:**
- NW = New order needed
- RE = Release (convert to MO)
- RI = Reschedule In (expedite)
- RO = Reschedule Out (delay)
- CN = Cancel
- IQ = Increase quantity
- RQ = Reduce quantity

---

## Generation and Updates

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| GETY | string | Generation Reference | How proposal was generated |
| NUAU | integer | Number of Auto Updates | Times auto-updated by MRP (max: 99999) |

---

## Quantities

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PPQT | number | Planned Quantity | Proposed production quantity |
| ORQA | number | Ordered Qty (Alt U/M) | Quantity in alternate unit |
| MAUN | string | Manufacturing U/M | Unit of measure for manufacturing |

---

## Dates and Times

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RELD | integer | Release Date | Suggested date to release to MO (YYYYMMDD) |
| STDT | integer | Start Date | Planned start date (YYYYMMDD) |
| FIDT | integer | Finish Date | Planned finish date (YYYYMMDD) |
| MSTI | integer | Start Time | Planned start time (HHMM) |
| MFTI | integer | Finish Time | Planned finish time (HHMM) |
| PLDT | integer | Planning Date | MRP planning date (YYYYMMDD) |

---

## Planning and Scheduling

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RESP | string | Responsible | Person/planner responsible |
| PRIP | string | Priority | Order priority code |
| PLGR | string | Work Center | Primary work center/planning group |
| WCLN | string | Production Line | Manufacturing production line |
| PRDY | number | Production Days | Number of production days required |
| LTRE | integer | Share of Lead Time | Percentage of total lead time (max: 999) |

---

## Reference Orders (Critical for Linking)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RORC | integer | Reference Order Category | Type of originating demand (max: 9) |
| RORN | string | Reference Order Number | Originating order number |
| RORL | integer | Reference Order Line | Line in originating order (max: 999999) |
| RORX | integer | Reference Line Suffix | Line suffix (max: 999) |
| RORH | string | Reference Order Header | Reference to header order |

**RORC Values:**
- 3 = Customer Order (demand from OOLINE)
- 2 = Manufacturing Order (component for parent MO)
- 4 = Distribution Order
- 5 = MRP Proposal (multi-level planning)

---

## Hierarchy

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PLLO | integer | Proposal Overlying Level | Parent MOP one level up (max: 9999999) |
| PLHL | integer | Proposal Highest Level | Top-level MOP in structure (max: 9999999) |
| ORDP | integer | Order Dependent | Is order-dependent proposal (max: 9) |

**Order Dependent Flag:**
- 0 = Stock replenishment
- 1 = Order-dependent (tied to specific demand)

---

## Routing and Structure

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AOID | string | Alternative Routing | Alternative routing ID |
| ORTY | string | Order Type | Manufacturing order type to use |
| STRT | string | Structure Type | Product structure type |
| ECVE | string | Revision Number | Engineering change revision |

---

## Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CFIN | integer | Configuration Number | Configuration instance (max: 9999999999) |
| ECVS | integer | Simulation Round | Configuration simulation (max: 999) |
| HDPR | string | Main Product | Main product in configuration |
| OPTZ | string | Option Z | Configuration option Z |
| OPTX | string | Option X | Configuration option X |
| OPTY | string | Option Y | Configuration option Y |

---

## Attributes

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATNR | integer | Attribute Number | Attribute model instance ID |

---

## Project Management

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PROJ | string | Project Number | Project identifier |
| ELNO | string | Project Element | Project element/WBS code |

---

## Material Availability

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MSPM | integer | Material Shortage PO/MO | Material shortage indicator (max: 99) |
| TSPM | integer | Tool Shortage | Tool shortage indicator (max: 9) |
| MSCD | integer | Material Shortage Check Date | Date of last shortage check (YYYYMMDD) |

**MSPM Values:**
- 0 = No shortage
- 1 = Shortage exists (purchased items)
- 2 = Shortage exists (manufactured items)
- 3 = Shortage both types

---

## Shop Workload Balancing (SWB)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TSDA | integer | SWB Timestamp Date | Shop workload balance date (YYYYMMDD) |
| TSTE | integer | SWB Timestamp Time | Shop workload balance time (HHMMSS) |
| PRAP | integer | SWB Processed | Processed by SWB (max: 9) |
| ACHD | integer | SWB Change Date | Last SWB change date (YYYYMMDD) |
| ACHT | integer | SWB Change Time | Last SWB change time (HHMM) |

---

## Subnetwork

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SUBN | integer | Sub-Network Mark | Part of subnetwork (max: 9) |
| SUBD | integer | Subnetwork Due Date | Subnetwork due date (YYYYMMDD) |
| CLGP | integer | Color Group | Visual grouping for planning (max: 99) |
| NTWP | integer | External Network Priority | Network priority (max: 99) |

---

## Rescheduling

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PULD | integer | Pull-up Delayed Orders | Pull forward delayed orders (max: 9) |
| PULN | integer | Pull-up Early Orders | Pull forward early orders (max: 9) |
| RIFD | integer | Reschedule In Filter Date | Filter for expedite messages (YYYYMMDD) |
| ROFD | integer | Reschedule Out Filter Date | Filter for delay messages (YYYYMMDD) |

---

## Process Manufacturing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PCCO | integer | Process Code | Process manufacturing code (max: 9) |
| MFPC | string | Process | Process identifier |

---

## Schedule

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SCHN | integer | Schedule Number | Supply chain schedule reference |

---

## Planning Policy

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PLCD | string | Planning Policy | Planning policy code |
| GRTI | string | Group Technology Class | GT classification |

---

## Configuration Release

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RLSR | number | Released Config Orders | Released quantity for configuration |

---

## Warning Messages

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MSG1 | string | Warning Message 1 | Planning warning or error 1 |
| MSG2 | string | Warning Message 2 | Planning warning or error 2 |
| MSG3 | string | Warning Message 3 | Planning warning or error 3 |
| MSG4 | string | Warning Message 4 | Planning warning or error 4 |

**Common Warning Messages:**
- Material shortage detected
- Capacity overload
- Routing not found
- BOM missing/invalid
- Lead time insufficient

---

## Consolidation

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CNSR | string | Consolidation Responsible | Responsible for order consolidation |

---

## External References

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| EXRN | string | External Reference Number | External order/reference number |
| EXD2 | string | External Reference Desc | Description of external reference |

---

## Sequencing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SQCO | string | Sequencing Code | Sequencing code for production |

---

## Document Identity

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DTID | integer | Document Identity | Document reference ID |
| PGNM | string | Program Name | Program that created the proposal |

---

## M3 Audit Fields

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RGDT | integer | Entry Date | Record creation date (YYYYMMDD) |
| RGTM | integer | Entry Time | Record creation time (HHMMSS) |
| LMDT | integer | Change Date | Last modification date (YYYYMMDD) |
| CHNO | integer | Change Number | Sequential change counter (max: 999) |
| CHID | string | Changed By | User who last modified |
| LMTS | integer | Timestamp | Last modification timestamp (microseconds) |

---

## Data Lake Metadata

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| accountingEntity | string | Accounting Entity | Record accounting entity |
| variationNumber | integer | Variation Number | Record modification sequence |
| timestamp | string (datetime) | Modification Timestamp | Record modification time (ISO format) |
| deleted | boolean | Is Deleted | Record deletion flag (STRING: "true"/"false") |
| archived | boolean | Is Archived | Record archive flag |

---

## Field Count by Category

| Category | Count | Description |
|----------|-------|-------------|
| Core Identifiers | 7 | Basic MOP identification |
| Status | 2 | Proposal and MO status |
| Action Messages | 1 | MRP action messages |
| Quantities | 3 | Planned quantities |
| Dates/Times | 6 | Planning dates |
| Planning/Scheduling | 6 | Planning parameters |
| Reference Orders | 5 | **Critical for MOP→CO linking** |
| Hierarchy | 3 | Multi-level proposal relationships |
| Configuration | 7 | Product configuration |
| Material Availability | 3 | Shortage indicators |
| SWB | 5 | Shop workload balancing |
| Subnetwork | 4 | Network planning |
| Rescheduling | 4 | Reschedule parameters |
| Warning Messages | 4 | Planning warnings |
| Project | 2 | Project references |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **91** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + FACI + PLPN + PLPS

### Foreign Keys / Relationships
- **To Customer Orders**: RORC=3, RORN=ORNO, RORL=PONR, RORX=POSX
- **To Parent MOP**: RORC=5, RORN references parent PLPN
- **To Manufacturing Orders**: When converted, becomes MWOHED record
- **To Item/Product**: ITNO, PRNO
- **To Facility**: FACI
- **To Warehouse**: WHLO
- **To Work Center**: PLGR

### Multi-Level Hierarchy
- **Top Level**: PLHL (highest proposal number)
- **Parent Level**: PLLO (overlying proposal)
- **Current Level**: PLPN + PLPS
- **Dependency**: ORDP (order-dependent flag)

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

### Most Commonly Used Fields
1. PLPN, PLPS, PRNO, ITNO (identification)
2. PSTS, ACTP (status and actions)
3. PPQT (planned quantity)
4. RELD, STDT, FIDT (dates)
5. RORC, RORN, RORL, RORX (demand links)
6. MSPM (material shortage)
7. MSG1-MSG4 (warnings)
8. LMDT (change tracking)

---

## MOP Lifecycle

```
Created by MRP → PSTS=10 (New)
                    ↓
                 PSTS=20 (Acknowledged)
                    ↓
              Convert to MO
                    ↓
                 PSTS=30 (Converted)
                    ↓
           (MOP record remains for history)
```

---

## Common Query Patterns

### Find MOPs for a Customer Order Line
```sql
SELECT * FROM MMOPLP
WHERE RORC = 3
  AND RORN = 'CO12345'
  AND RORL = 1
  AND PSTS IN ('10', '20')  -- Active proposals only
  AND deleted = 'false'
```

### Active MOPs with Action Messages
```sql
SELECT * FROM MMOPLP
WHERE PSTS IN ('10', '20')
  AND ACTP IS NOT NULL
  AND ACTP != ''
  AND deleted = 'false'
ORDER BY RELD, PLPN
```

### MOPs with Material Shortages
```sql
SELECT * FROM MMOPLP
WHERE MSPM > 0  -- Has shortage
  AND PSTS IN ('10', '20')
  AND deleted = 'false'
ORDER BY RELD
```

### Multi-Level MOP Structure
```sql
-- Get full hierarchy
SELECT child.*, parent.*
FROM MMOPLP child
LEFT JOIN MMOPLP parent
  ON parent.PLPN = child.PLLO
  AND parent.FACI = child.FACI
WHERE child.PLHL = 123456  -- Top-level proposal
  AND child.deleted = 'false'
ORDER BY child.ORDP DESC, child.PLPN
```

### MOPs Ready to Release
```sql
SELECT * FROM MMOPLP
WHERE PSTS = '10'
  AND ACTP = 'RE'  -- Release action
  AND RELD <= 20260130  -- Release date passed
  AND MSPM = 0  -- No material shortage
  AND deleted = 'false'
ORDER BY RELD, PLPN
```

---

## Relationship to Manufacturing Orders

When a MOP (MMOPLP) is converted to an MO (MWOHED):

1. **PSTS changes** from '10' or '20' to '30' (Converted)
2. **New MO created** in MWOHED with:
   - MFNO = new MO number
   - GETP = 4 (From proposal)
   - Links back via RORC/RORN to same demand
3. **MOP record retained** for planning history

---

## Critical Planning Fields

### For Supply Chain Analysis
- **RORC, RORN, RORL, RORX**: Link to demand
- **PPQT**: Quantity to produce
- **RELD, STDT, FIDT**: Time fence
- **MSPM**: Material availability
- **ACTP**: Required action

### For Capacity Planning
- **PLGR, WCLN**: Resource requirements
- **STDT, FIDT**: Time requirements
- **PRDY**: Duration

### For MRP Analysis
- **ACTP**: What action is needed
- **MSG1-MSG4**: Planning warnings
- **NUAU**: Stability (low = stable, high = unstable demand)

---

## Action Message Priority

When multiple action messages exist, typical priority:
1. **RE** (Release) - Create MO now
2. **RI** (Reschedule In) - Expedite
3. **IQ** (Increase Quantity) - Increase production
4. **RQ** (Reduce Quantity) - Reduce production
5. **RO** (Reschedule Out) - Delay
6. **CN** (Cancel) - Not needed

---

## Integration Points

### MRP Flow
```
Customer Order (OOLINE)
    ↓
Demand Planning
    ↓
MOP Created (MMOPLP) - RORC=3, links to CO
    ↓
Material Check (MSPM populated)
    ↓
Capacity Check (SWB processing)
    ↓
Action Message (ACTP)
    ↓
Convert to MO (MWOHED)
```

### Multi-Level Planning
```
Top-Level MOP (PLHL)
    └─ Level 1 MOP (PLLO points to parent)
        └─ Level 2 MOP (PLLO points to level 1)
```
