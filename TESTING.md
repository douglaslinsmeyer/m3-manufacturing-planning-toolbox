# Testing the Compass Data Fabric Integration

## Prerequisites

1. ✅ All Docker services running
2. ✅ Authenticated in the application (logged in via http://localhost:3000)
3. ✅ OAuth redirect URL registered in Infor ION API

## Testing Data Refresh

### Step 1: Login to the Application

1. Open http://localhost:3000
2. Select **TRN** environment
3. Click "Sign In with M3"
4. Authenticate with your M3 credentials
5. You should land on the dashboard

### Step 2: Trigger Data Refresh

On the dashboard, click the **"Refresh Data"** button.

This will:
1. Query Compass Data Fabric for:
   - Customer Order Lines (OOLINE)
   - Manufacturing Orders (MWOHED)
   - Planned Manufacturing Orders (MMOPLP)
2. Parse and transform the data
3. Insert into PostgreSQL with JSONB attributes
4. Update the unified `production_orders` view

### Step 3: Monitor Progress

Watch the backend logs in real-time:

```bash
docker-compose logs -f backend
```

You should see:
```
Starting data refresh from M3...
Refreshing customer order lines...
Using last sync date: 20200101
Submitting Compass query for CO lines...
Query completed successfully. JobID: xxx, Records: 1234
Parsing CO line results...
Received 1234 CO line records
Inserting 1234 CO line records into database...
CO lines refresh completed
Refreshing manufacturing orders...
...
Data refresh completed successfully
```

### Step 4: Verify Data in Database

```bash
# Connect to PostgreSQL
docker-compose exec db psql -U postgres -d m3_planning

# Check record counts
SELECT COUNT(*) FROM customer_order_lines;
SELECT COUNT(*) FROM manufacturing_orders;
SELECT COUNT(*) FROM planned_manufacturing_orders;
SELECT COUNT(*) FROM production_orders;

# View sample CO line with attributes
SELECT
  order_number,
  line_number,
  item_number,
  status,
  attributes->'builtin_string'->>'ATV6' as color,
  attributes->'user_defined_alpha'->>'UCA1' as custom_field
FROM customer_order_lines
LIMIT 5;

# View unified production orders
SELECT
  order_number,
  order_type,
  item_number,
  planned_start_date,
  planned_finish_date,
  status
FROM production_orders
ORDER BY planned_start_date
LIMIT 10;

# Check incremental load tracking
SELECT version, applied_at FROM schema_migrations;
SELECT MAX(lmdt) as last_sync_date FROM customer_order_lines;
```

## Understanding the Data Flow

### Compass Query Execution

1. **Submit Query**:
   - POST `/DATAFABRIC/compass/v2/jobs/`
   - Returns `jobId`

2. **Poll Status**:
   - GET `/DATAFABRIC/compass/v2/jobs/{jobId}/status/`
   - Polls every 2 seconds
   - Waits for status: `completed`

3. **Fetch Results**:
   - GET `/DATAFABRIC/compass/v2/jobs/{jobId}/result/?offset=0&limit=100000`
   - Returns JSON with rows and columns

### Data Transformation

```
Compass JSON → Parse → Build JSONB → Database Record
```

Example transformation for CO line:

```json
// Compass result
{
  "ORNO": "CO12345",
  "PONR": 1,
  "ATV6": "Color:Red",
  "UCA1": "CustomValue",
  "DIP1": 5.0
}

// Transformed to
{
  "order_number": "CO12345",
  "line_number": "1",
  "attributes": {
    "builtin_string": {
      "ATV6": "Color:Red"
    },
    "user_defined_alpha": {
      "UCA1": "CustomValue"
    },
    "discounts": {
      "percentages": {
        "DIP1": 5.0
      }
    }
  }
}
```

## Incremental Loading

The system automatically performs **incremental loads** based on M3's `LMDT` (Last Modified Date) field:

### First Refresh
- Uses `LMDT >= 20200101` (January 1, 2020)
- Loads all historical data

### Subsequent Refreshes
- Checks: `SELECT MAX(lmdt) FROM customer_order_lines`
- Uses that date: `LMDT >= {last_sync_date}`
- Only loads changed/new records

This ensures efficient refresh operations on subsequent runs.

## Testing Queries

### Test Compass Client Directly

You can test the Compass queries using the Infor MCP server:

```bash
# From the Infor MCP server project
cd ~/Projects/infor-mcp-server

# Test CO line query
echo 'SELECT ORNO, PONR, ITNO, ORST FROM OOLINE WHERE deleted = '"'"'false'"'"' AND LMDT >= 20240101 LIMIT 10' | \
  npx ts-node test-query.ts
```

### Sample Queries for Verification

```sql
-- Find CO lines linked to MOs
SELECT
  co.order_number,
  co.line_number,
  co.item_number,
  mo.mo_number,
  mo.status as mo_status
FROM customer_order_lines co
JOIN manufacturing_orders mo
  ON mo.rorc = 3
  AND mo.rorn = co.order_number
  AND mo.rorl = CAST(co.line_number AS INTEGER)
WHERE co.rorc IS NOT NULL;

-- Find production timeline
SELECT
  order_number,
  order_type,
  planned_start_date,
  planned_finish_date,
  status
FROM production_orders
WHERE planned_start_date >= CURRENT_DATE
ORDER BY planned_start_date, order_type;

-- Find MOPs linked to CO lines
SELECT
  mop.mop_number,
  mop.status,
  co.order_number,
  co.line_number
FROM planned_manufacturing_orders mop
JOIN customer_order_lines co
  ON mop.rorc = 3
  AND mop.rorn = co.order_number
  AND mop.rorl = CAST(co.line_number AS INTEGER);
```

## Troubleshooting

### No data returned from Compass

Check:
1. OAuth token is valid (check session)
2. Compass Base URL is correct in .env
3. `deleted = 'false'` is using string comparison (not boolean)
4. LMDT filter isn't too restrictive

### Database insert failures

Check:
1. Migration 002 applied successfully
2. JSONB columns exist: `SELECT column_name FROM information_schema.columns WHERE table_name = 'customer_order_lines' AND column_name = 'attributes';`
3. Foreign key constraints aren't blocking

### Query timeout

For large datasets:
1. Compass has a 5-minute HTTP timeout
2. Consider splitting by date ranges
3. Use NATS for truly async processing (future enhancement)

## What's Working Now

✅ **Compass Data Fabric Integration**:
- HTTP client with OAuth token injection
- Query submission and status polling
- Result fetching with pagination support
- Automatic retry and error handling

✅ **Query Builder**:
- CO line query (OOLINE) with all attributes
- MO query (MWOHED) with hierarchy
- MOP query (MMOPLP) with messages

✅ **Data Parser**:
- JSON to structured Go types
- JSONB attribute builder for flexible fields
- M3 date/time parsing (YYYYMMDD, HHMM formats)
- Safe type conversions

✅ **Database Layer**:
- Batch insert for efficiency
- Upsert logic (INSERT ON CONFLICT UPDATE)
- Incremental load tracking via LMDT
- Unified production_orders view updates

✅ **API Integration**:
- `/api/snapshot/refresh` endpoint working
- Async refresh (runs in goroutine)
- Session-based Compass client per user

## Next Steps

1. **Test with real M3 data** - Click "Refresh Data" on dashboard
2. **Add NATS workers** - Move async processing to NATS for better scalability
3. **Implement progress tracking** - Show real-time progress in UI
4. **Add analysis engine** - Detect inconsistencies between MOs and CO delivery dates
5. **Build frontend views** - Display production orders, inconsistencies, etc.

The foundation is complete and ready for testing with real M3 data!
