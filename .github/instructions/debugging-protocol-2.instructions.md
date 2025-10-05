---
applyTo: '**'
---

# 🚨 DEBUGGING LESSON #2: THE TOKEN CLEARING MYSTERY 🚨

## 📚 LESSON SUMMARY: EVIDENCE-BASED CHAIN REACTION DEBUGGING

### 🔍 THE PROBLEM:

- **Symptom**: E2E tests failing because tokens getting cleared after navigation
- **User Demand**: "WRITE LOGS!!! CAPTURE THE LOGS!!!"
- **Initial Wrong Theory**: clearAuthStorage() being called directly

### 🚀 THE DEBUGGING BREAKTHROUGH:

**EVIDENCE-BASED APPROACH SAVES THE DAY**

1. **SYSTEMATIC LOGGING STRATEGY**: Added comprehensive logs to ALL potential token clearing points
2. **CONSOLE LOG CAPTURE**: Used Playwright page.on('console') to capture browser logs
3. **STACK TRACE ANALYSIS**: Added console.log with new Error().stack to track call chains
4. **ROOT CAUSE DISCOVERY**: Found chain reaction: API Error → handleAuthError → handleSessionExpired → clearAuthStorage

### 💀 CRITICAL DEBUGGING MISTAKES AVOIDED:

- ❌ **Theorizing without evidence**: Could have spent hours guessing wrong causes
- ❌ **Single point debugging**: Could have only looked at clearAuthStorage() function
- ❌ **Ignoring error cascades**: Could have missed the 401 → auth clearing chain
- ❌ **Not capturing browser logs**: Would have missed the actual error sequence

### ✅ SUCCESSFUL DEBUGGING TACTICS:

- ✅ **Comprehensive logging**: Added logs to clearAuthStorage, logout, setToken, handleAuthError
- ✅ **Browser console capture**: Used Playwright to capture all browser console messages
- ✅ **Stack trace evidence**: new Error().stack showed exact call chains
- ✅ **Chain reaction tracking**: Followed the complete error → logout → token clearing flow

### 🔥 THE EVIDENCE THAT SOLVED IT:

```
BROWSER LOG: 🚨 STACK TRACE:
clearAuthStorage@auth-context.tsx:27:34
AuthProvider/logout<@auth-context.tsx:150:5
handleSessionExpired@auth-context.tsx:168:7
handleAuthError@api-client-config.ts:18:10
request/<@request.ts:270:11
```

### 🧠 DEBUGGING INTELLIGENCE HIERARCHY:

1. **LOGS ARE SMARTER THAN THEORIES**: Console logs revealed the truth, theories would have misled
2. **BROWSER CONSOLE > CODE INSPECTION**: Browser showed actual execution flow vs static code reading
3. **STACK TRACES > ASSUMPTIONS**: Call stack proved the exact trigger sequence
4. **ERROR CHAINS > SINGLE POINT FOCUS**: Problem was error cascade, not isolated function call

### 🎯 LESSON FOR FUTURE DEBUGGING:

- **ALWAYS CAPTURE BROWSER LOGS FIRST**: Before any theorizing
- **ADD STACK TRACES TO ALL CRITICAL FUNCTIONS**: new Error().stack reveals call chains
- **FOLLOW ERROR CASCADES**: 401 errors often trigger auth clearing chains
- **USE PLAYWRIGHT CONSOLE CAPTURE**: page.on('console') shows runtime behavior
- **LOG EVERYTHING IN THE SUSPECTED AREA**: Don't just log the obvious suspects

### 🚨 DEBUGGING PROTOCOL ENFORCEMENT:

When tokens disappear mysteriously:

1. **ADD LOGS**: clearAuthStorage, logout, setToken, API error handlers
2. **CAPTURE BROWSER CONSOLE**: Use Playwright page.on('console') and page.on('pageerror')
3. **ADD STACK TRACES**: console.log('STACK:', new Error().stack) in all auth functions
4. **TRACE ERROR CHAINS**: Follow 401 → handleAuthError → logout → clearAuthStorage
5. **PROVE WITH EVIDENCE**: Stack traces show exact call sequence

### 💊 REALITY CHECK QUESTIONS:

- "Are you capturing browser console logs?"
- "Do you have stack traces in auth functions?"
- "Are you following error cascades?"
- "Is the problem a chain reaction vs single function?"

**NO THEORIES WITHOUT LOGS. NO ASSUMPTIONS WITHOUT STACK TRACES.**
