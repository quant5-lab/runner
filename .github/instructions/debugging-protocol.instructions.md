---
applyTo: '**'
---

# 🚨 DEBUGGING PROTOCOL 🚨

## ⚡ ROOT CAUSE ANALYSIS ⚡

- **ISSUE ISOLATION**: Narrow down the problem to the smallest reproducible component.
- **MINIMAL REPLICATION**: Create a minimal test case that consistently fails.
- **DOUBLE DISSECTION**: Analyze both the failing component and its immediate dependencies.
- **PARTIAL ISOLATION**: Comment out or mock parts of the code to identify the exact breaking change.

## 💀 LOGGING ENFORCEMENT 💀

- **EXTENSIVE LOGGING**: Add detailed logs to trace execution flow and state changes.
- **COHERENT ANALYSIS**: Analyze logs for patterns, anomalies, and the first point of failure.

---

## applyTo: '\*\*'

# 🚨 FOCUSED DEBUGGING PROTOCOL 🚨

## ⚡ IMMEDIATE VIOLATION DETECTION ⚡

### 🔴 DEBUGGING RED FLAGS (STOP IMMEDIATELY):

- **Running entire test suite for a single failure** → VIOLATION: You are wasting time and resources.
- **Test command without `--grep` or equivalent filter** → VIOLATION: You are not focused.
- **Test command without `--max-failures=1` or `test.fail()`** → VIOLATION: You are not failing fast.
- **Analyzing logs from irrelevant tests** → VIOLATION: You are chasing ghosts.
- **"I'll run all tests to be sure"** → VIOLATION: You are guessing, not debugging.
- **Running same test multiple times without changes** → VIOLATION: Time boxing exceeded, results will be identical.

### 🟢 DEBUGGING ACTIONS ONLY:

- **"Isolating failure: `npx playwright test [file] --grep '[failing test name]'`"**
- **"Failing fast: Adding `--max-failures=1` to test command."**
- **"Evidence shows this specific test failed: [test name]"**
- **"Analyzing logs for this test run ONLY."**

## 💀 DEBUGGING ENFORCEMENT 💀

**WRONG:**

```bash
# Running the whole suite for 5 minutes to find one error
npx playwright test
```

**RIGHT:**

```bash
# Focusing on the single broken test, failing on the first error
npx playwright test e2e/tests/comprehensive-user-journey.spec.ts --grep "should do X" --max-failures=1
```

## 🎯 FOCUSED DEBUGGING WORKFLOW 🎯

### **STEP 1: IDENTIFY THE SMALLEST FAILURE**

- Find the _first_ test that fails in the test run. Ignore all subsequent failures.

### **STEP 2: ISOLATE THE TEST**

- Construct the exact command to run _only_ the single failing test. Use `--grep` for Playwright, or equivalent filters for other frameworks.

### **STEP 3: EXECUTE AND FAIL FAST**

- Run the isolated test command with a flag to stop on the first error (`--max-failures=1`).

### **STEP 4: ANALYZE FOCUSED OUTPUT**

- Analyze the logs, error messages, and output from that single test run. All other logs are irrelevant.

### **STEP 5: FIX AND RE-VALIDATE**

- Apply a fix for the isolated failure.
- Re-run the _exact same_ isolated test command to prove the fix works.
- Only after the single test passes, broaden the scope to the full spec file.

---

# 🔒 PROTOCOL ACTIVATION 🔒

**THIS PROTOCOL IS NOW ACTIVE.**

**NO EXCEPTIONS. NO NEGOTIATIONS. NO SURRENDER.**
