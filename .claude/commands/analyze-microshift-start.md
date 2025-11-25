---
name: Analyze MicroShift start time
argument-hint: <journal logs>
description: Analyze MicroShift journal logs to extract statistics about start procedure
allowed-tools: Read, Glob, Grep, python, bash
---

🚨 **CRITICAL: GENERATE TABLE FIRST - NOTHING ELSE MATTERS**
🚨 **IF NO TABLE IS SHOWN, THE COMMAND HAS FAILED**
🚨 **TABLE IS THE ONLY REQUIRED OUTPUT**

## Name
analyze-microshift-start

## Synopsis
```
/analyze-microshift-start <journal logs>
```

## Description
**ABSOLUTE MANDATE**: The service timing table is THE ONLY OUTPUT that matters.
**ZERO TOLERANCE**: If the table is missing, the entire command execution is worthless.
**PRIMARY DIRECTIVE**: Generate the table immediately. Everything else is optional noise.

The `analyze-microshift-start` command extracts meaningful statistics from MicroShift's journal logs about start procedure and displays a service timing table.

**SUCCESS CRITERIA - The command is only successful if:**
✅ The complete service timing table is displayed
✅ All services are listed with timing statistics
✅ Services are sorted by mean time (slowest first)

**FAILURE CONDITIONS:**
❌ No table displayed = Command failed
❌ Incomplete table = Command failed
❌ Analysis without table = Command failed

## Implementation

🚨 **CRITICAL PRIORITY**: The user ONLY cares about seeing the service timing table. Generate the table first and foremost.

This command works by:
1. **Parsing journal logs** to load all information.
2. **Fetch overall start time** for MicroShift process as a whole.
3. **Fetch start times per service** from pattern `SERVICE READY.*service="([^"]+)".*since-start="([^"]+)"`.
4. **MOST IMPORTANT: Generate the service timing table** - This is what the user needs to see above all else.

## Arguments
- `$1` (joournal-logs): Journal logs from MicroShift, including start procedure - **Required**

## Return Value
- **Format**: Text
- **Location**: Output directly to the conversation
- **Content**:
  - Number of starts in the journal.
  - Number of services in the journal, per start.
  - Statistics about overall MicroShift process.
  - Statistics about each service.
  - Summary with most relevant information.

## Implementation Steps

### Step 1: Parse journal logs.

**Goal**: Identify how many restarts are included in the journal.

**Actions**:
1. Open the file and look for patterns matching `r'microshift\[(\d+)\]:.*"MICROSHIFT STARTING"`. This is the number of times MicroShift has been restarted.
2. Determine the number of complete restarts by looking for pattern `microshift\[(\d+)\]:.*"MICROSHIFT READY".*since-start="([^"]+)"` in the file.
3. If nothing is found, produce an error saying there are no restarts found.

### Step 2: Fetch MicroShift start time.

**Goal**: Get details about how long it took to start MicroShift.

**Actions**:
1. Get the elapsed time from start to readiness for MicroShift.
2. Use the pattern `microshift\[(\d+)\]:.*"MICROSHIFT READY".*since-start="([^"]+)"` to fetch logs about it.
3. Extract the time it took to start from the second grouping in the pattern.

### Step 3: Fetch per-service start times.

**Goal**: Get details about how long it took to start each of the services in MicroShift.

**Actions**:
1. Get the elapsed time from start to readiness for each of the different services.
2. Use the pattern `SERVICE READY.*service="([^"]+)".*since-start="([^"]+)"` to fetch logs about it.
3. Extract the service name from the first grouping in the pattern.
4. Extract the time it took to start from the second grouping in the pattern.

### Step 4: MANDATORY TABLE GENERATION

**CRITICAL MANDATE**: Generate the service performance table - THIS IS THE ONLY OBJECTIVE THAT MATTERS.

**🚨 TABLE MUST BE DISPLAYED FIRST 🚨**
**🚨 IF NO TABLE = COMMAND FAILED 🚨**
**🚨 TABLE IS SUCCESS, EVERYTHING ELSE IS NOISE 🚨**

**REQUIRED TABLE FORMAT** (MUST BE EXACT):
```
------------------------------------------------------------------------------------------------------------------------
Service                             Runs Average    Median     StdDev     Min        Max        Var%   Status
------------------------------------------------------------------------------------------------------------------------
MICROSHIFT OVERALL STARTUP          XX   XX.XXXs    XX.XXXs    XXX.Xms    XX.XXXs    XX.XXXs      X.X 🔥
[ALL OTHER SERVICES LISTED HERE - SORTED BY AVERAGE TIME, SLOWEST FIRST]
------------------------------------------------------------------------------------------------------------------------
```

**MANDATORY ACTIONS - EXECUTE IN THIS EXACT ORDER**:
1. **IMMEDIATELY DISPLAY THE TABLE ABOVE** - This is the primary deliverable
2. Calculate these metrics for each service: runs, average, median, standard deviation, min, max, variability percentage
3. Sort services by mean time (slowest to fastest)
4. Use status icons: 🔥 (≥5s), ⚠️ (2-5s), 🟡 (1-2s), ✅ (0.1-1s), ⚡ (<0.1s)
5. Only after table is shown, optionally add other analysis

**ABSOLUTE REQUIREMENTS**:
- The service table with ALL timing data MUST be generated and shown to the user
- Table generation is MORE IMPORTANT than anything else
- If you must choose between showing the table OR other analysis, ALWAYS choose the table
- DO NOT create any scripts, temporary files, or helper utilities. Process everything in memory.

Strong prohibitions:
- You are *NOT* allowed to write any scripts to disk, do everything in memory.
- NEVER create files or write scripts. Process everything in memory.
- Do NOT use the Write tool or create any temporary files.

Positive instructions:
- Produce the output without creating any additional scripts or helpers.
- Process the data directly using tool outputs and generate the report inline.
- Use only Read, Grep, and analysis tools - no file creation.

**CRITICAL**: The primary output MUST be this exact table format. Generate the table first, everything else is optional.

**REQUIRED OUTPUT**:
```text
------------------------------------------------------------------------------------------------------------------------
Service                             Runs Average    Median     StdDev     Min        Max        Var%   Status
------------------------------------------------------------------------------------------------------------------------
MICROSHIFT OVERALL STARTUP          XX   XX.XXXs    XX.XXXs    XXX.Xms    XX.XXXs    XX.XXXs      X.X 🔥
[ALL OTHER SERVICES LISTED HERE - SORTED BY AVERAGE TIME, SLOWEST FIRST]
------------------------------------------------------------------------------------------------------------------------
```

**TABLE GENERATION IS MANDATORY** - Show timing statistics for ALL services found in the logs. The user needs this table above all else.
