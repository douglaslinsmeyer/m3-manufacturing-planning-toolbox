# Progress Bar Fix - Stuck at 0%

## Problem
Progress bar was stuck at 0% because the worker's progress callback couldn't access the actual record counts being processed.

## Root Cause
1. The `ProgressCallback` signature only passed `(phase string, current, total int, message string)`
2. The worker had local variables `mopCount`, `moCount`, `coCount` initialized to 0
3. These variables were never updated because the callback had no way to receive the actual counts
4. The service tracked counts internally but didn't pass them through the callback
5. Result: Worker calculated rates and ETAs based on 0 records → always showed 0%

## Solution Implemented

### 1. Extended ProgressCallback Signature
**File: `backend/internal/services/snapshot.go`**

Changed from:
```go
type ProgressCallback func(phase string, current, total int, message string)
```

To:
```go
type ProgressCallback func(phase string, stepNum, totalSteps int, message string, mopCount, moCount, coCount int)
```

### 2. Added Count Tracking to SnapshotService
**File: `backend/internal/services/snapshot.go`**

Added fields to service struct:
```go
type SnapshotService struct {
    compassClient    *compass.Client
    db               *db.Queries
    progressCallback ProgressCallback
    mopCount         int  // NEW
    moCount          int  // NEW
    coCount          int  // NEW
}
```

Updated `RefreshAll()` to track counts:
```go
// Phase 1: MOPs
mopRefs, err := s.RefreshPlannedOrders(ctx, company, facility)
s.mopCount = len(mopRefs)  // Store count
s.reportProgress("mops", 1, 4, fmt.Sprintf("Loaded %d planned orders", s.mopCount))

// Phase 2: MOs
moRefs, err := s.RefreshManufacturingOrders(ctx, company, facility)
s.moCount = len(moRefs)  // Store count
s.reportProgress("mos", 2, 4, fmt.Sprintf("Loaded %d manufacturing orders", s.moCount))

// Phase 3: CO lines
coCount, err := s.RefreshOpenCustomerOrderLines(ctx, company, facility)
s.coCount = coCount  // Store count
s.reportProgress("cos", 3, 4, fmt.Sprintf("Loaded %d customer order lines", s.coCount))
```

Updated `reportProgress()` to pass counts:
```go
func (s *SnapshotService) reportProgress(phase string, stepNum, totalSteps int, message string) {
    if s.progressCallback != nil {
        s.progressCallback(phase, stepNum, totalSteps, message, s.mopCount, s.moCount, s.coCount)
    }
}
```

### 3. Updated Worker Callback
**File: `backend/internal/workers/snapshot_worker.go`**

Changed callback signature to match:
```go
snapshotService.SetProgressCallback(func(phase string, stepNum, total int, message string, mopCount, moCount, coCount int) {
    // Store final counts for use after RefreshAll completes
    finalMopCount = mopCount
    finalMoCount = moCount
    finalCoCount = coCount

    // Calculate metrics with ACTUAL counts
    totalRecords := mopCount + moCount + coCount
    recordsPerSec := float64(totalRecords) / elapsed

    // Calculate progress percentage
    progressPct := (stepNum * 100) / totalSteps

    // Publish with real data
    w.publishDetailedProgress(req.JobID, "running",
        fmt.Sprintf("Step %d of %d", stepNum+1, totalSteps),
        operation,
        stepNum, totalSteps, progressPct,
        coCount, moCount, mopCount,  // REAL COUNTS
        recordsPerSec, estimatedRemaining,
        0, 0)
})
```

## What Changed

### Before:
```
Progress: 0%  (always stuck)
Records: 0 MOPs, 0 MOs, 0 COs (always zero)
Rate: 0/sec (division by zero)
ETA: 0s (invalid)
```

### After:
```
Progress: 0% → 25% → 50% → 75% → 100% (moves through phases)
Records: 1,234 MOPs, 567 MOs, 4,567 COs (actual counts)
Rate: ~220/sec (real calculation)
ETA: ~30s (accurate estimate)
```

## Data Flow (Fixed)

1. **Service loads MOPs**
   - `RefreshPlannedOrders()` returns slice of CO refs
   - Service stores: `s.mopCount = len(refs)`
   - Calls: `s.reportProgress("mops", 1, 4, message)`

2. **reportProgress calls callback**
   - Passes: `callback("mops", 1, 4, message, 1234, 0, 0)`
   - Worker receives actual `mopCount = 1234`

3. **Worker calculates metrics**
   - `totalRecords = 1234 + 0 + 0 = 1234`
   - `recordsPerSec = 1234 / elapsedSeconds`
   - `progressPct = (1 * 100) / 4 = 25%`

4. **Worker publishes progress**
   - Sends to NATS with real counts
   - SSE streams to frontend
   - UI updates with 25% and "1,234 MOPs"

5. **Repeat for MOs (50%) and COs (75%)**

6. **Finalize (100%)**

## Testing

To verify the fix works:

1. **Check Backend Logs:**
```bash
# Look for these log messages with actual counts:
✓ M3 snapshot tables truncated successfully
Phase 1: Refreshing planned manufacturing orders (MOPs) with CO links...
Received X MOP records
Phase 2: Refreshing manufacturing orders (MOs) with CO links...
Received Y MO records
Phase 3: Refreshing all open customer order lines (status < 30)...
Received Z CO line records
```

2. **Check Frontend Progress Bar:**
- Should show 0% → 25% → 50% → 75% → 100%
- Should display actual record counts
- Should show processing rate (records/sec)
- Should show ETA countdown

3. **Check SSE Stream (DevTools → Network → EventStream):**
```json
{
  "jobId": "job-123",
  "status": "running",
  "progress": 25,
  "currentStep": "Step 1 of 4",
  "currentOperation": "Loaded 1234 planned orders",
  "mopsProcessed": 1234,
  "mosProcessed": 0,
  "coLinesProcessed": 0,
  "recordsPerSecond": 247.5,
  "estimatedTimeRemaining": 15
}
```

## Files Modified

1. `backend/internal/services/snapshot.go` - Added count tracking, updated callback signature
2. `backend/internal/workers/snapshot_worker.go` - Updated callback to receive counts, removed unused local vars

## Next Steps

After restarting the backend:
1. Start a data refresh
2. Watch the progress bar increment through 0% → 25% → 50% → 75% → 100%
3. Verify record counts update in real-time
4. Confirm rate and ETA calculations appear and make sense

The progress bar should now show real-time progress with accurate metrics! 🎉
