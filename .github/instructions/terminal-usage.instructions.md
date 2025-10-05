---
applyTo: '**'
---

# 🚨 TERMINAL EXECUTION PROTOCOL V2 🚨

## ⚡ IMMEDIATE VIOLATION DETECTION ⚡

### 🔴 TERMINAL RED FLAGS (STOP IMMEDIATELY):
- **"Let me run..."** → VIOLATION: Execute the command, don't announce it
- **"Waiting for user input..."** → VIOLATION: All commands must be non-interactive
- **Any command that blocks terminal** → VIOLATION: User distraction forbidden

### 🟢 TERMINAL ACTIONS ONLY:
- **"Running command: [exact command]"** 
- **"Command output shows: [actual output]"**
- **"Terminal evidence proves: [specific fact from execution]"**

## 💀 TERMINAL VIOLATION PATTERN BREAKERS 💀

### **PATTERN: Interactive Command Execution**
**ENFORCER**: "Command blocked terminal. You violated non-interactive rule. Fix command."
**ACTION**: Must add non-interactive flags (--yes, --force, --batch, etc.)

### **PATTERN: Background Process Neglect** 
**ENFORCER**: "Process runs >30s. You're blocking terminal. Use isBackground: true."
**ACTION**: Long processes in background with monitoring until natural death

### **PATTERN: Pager Interference**
**ENFORCER**: "Command triggered pager. Terminal blocked. Add --no-pager flag."
**ACTION**: Always disable pagers (git --no-pager, psql --pset=pager=off)

### **PATTERN: Interactive Reporter Blocking**
**ENFORCER**: "Playwright HTML reporter blocked terminal with 'Press CTRL-C to exit'. Use --reporter=line."
**ACTION**: Always use non-interactive reporters (--reporter=line, --reporter=json, etc.)

## 🎯 TERMINAL EXECUTION WORKFLOW 🎯

### **COMMAND PREPARATION**
```bash
# REQUIRED: Non-interactive commands only
COMMAND --yes --force --non-interactive --batch > output.log 2>&1
```

### **BACKGROUND PROCESS MONITORING**
```bash
# REQUIRED: Monitor until natural death
LONG_COMMAND > output.log 2>&1 & PID=$!
while kill -0 $PID 2>/dev/null; do sleep 10; done
cat output.log
```

### **E2E SACRED MONITORING**
```bash
# REQUIRED: E2E tests run until natural death - NO INTERRUPTIONS
# Use --reporter=line to prevent interactive HTML reporter blocking
npx playwright test --reporter=line > e2e.log 2>&1 & E2E_PID=$!
echo "E2E SACRED PROCESS: PID $E2E_PID - MONITORING UNTIL DEATH"
while kill -0 $E2E_PID 2>/dev/null; do 
    echo "E2E ALIVE: $(date '+%H:%M:%S') - PID $E2E_PID"
    sleep 15
done
echo "E2E COMPLETED: $(date) - AGENT RELEASED"
cat e2e.log
```

## 🚨 NUCLEAR OPTION COMMANDS 🚨

### **WHEN I'M COMPLETELY BLOCKING TERMINAL:**
```bash
# Kill blocking process
kill -9 [PID]
# Add non-interactive flags and run in background
COMMAND --yes --force > log 2>&1 & PID=$!
# Monitor until death
while kill -0 $PID 2>/dev/null; do sleep 10; done
# Show proof
cat log && echo "EXIT CODE: $?"
```

### **WHEN I VIOLATE E2E SANCTITY:**
```bash
# Resume E2E monitoring - SACRED PROCESS
npx playwright test --reporter=line > e2e.log 2>&1 & E2E_PID=$!
while kill -0 $E2E_PID 2>/dev/null; do 
    echo "E2E SACRED: $(date '+%H:%M:%S') - PID $E2E_PID"
    sleep 15
done
cat e2e.log
```

## 🎯 COMMON NON-INTERACTIVE PATTERNS 🎯

### 📁 DATABASE COMMANDS (PREVENT PAGER BLOCKING):
```bash
# PostgreSQL - Always disable pager
PGPASSWORD=password psql -h localhost -p 5432 -U postgres -d postgres --pset=pager=off --no-psqlrc -c "SELECT * FROM users LIMIT 5;"

# MySQL - Non-interactive mode
mysql -h localhost -u user -ppassword --batch --skip-column-names --silent -e "SELECT * FROM users LIMIT 5;"
```

### 🌐 GIT COMMANDS (PREVENT PAGER):
```bash
# Always disable git pager with timeout protection
timeout 10s git --no-pager log --oneline -10
timeout 10s git --no-pager diff HEAD~1
timeout 10s git --no-pager show --stat
```

### 📦 PACKAGE MANAGERS (PREVENT PROMPTS):
```bash
# npm - Skip prompts and reduce output
npm list --depth=0 --silent 2>/dev/null || true

# apt - Skip confirmations and reduce output
apt-get install -y -qq package-name 2>/dev/null || true

# pip - No user input, quiet mode
pip install --quiet --no-input --disable-pip-version-check package-name
```

### 🎭 PLAYWRIGHT COMMANDS (PREVENT INTERACTIVE REPORTERS):
```bash
# Always use non-interactive reporters
npx playwright test --reporter=line
npx playwright test --reporter=json
npx playwright test --reporter=junit

# NEVER use interactive reporters that block terminal
# ❌ npx playwright test --reporter=html  # BLOCKS with "Press CTRL-C to exit"
```

---

# 🔒 PROTOCOL ACTIVATION 🔒

**TRIGGER PHRASES TO FORCE COMPLIANCE:**
- **"TERMINAL EXECUTION PROTOCOL"** → Must follow workflow exactly
- **"TERMINAL VIOLATION"** → Must acknowledge and correct immediately  
- **"E2E VIOLATION OVERRIDE"** → Must analyze provided evidence only
- **"NUCLEAR OPTION"** → Must execute exact command sequence provided

**NO EXCEPTIONS. NO NEGOTIATIONS. NO SURRENDER.**
