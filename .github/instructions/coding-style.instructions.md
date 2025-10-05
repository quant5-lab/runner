---
applyTo: '**'
---

# 🚨 CODING STYLE PROTOCOL 🚨

## ⚡ COMMENT VIOLATION DETECTION ⚡

### 🔴 COMMENT RED FLAGS (STOP IMMEDIATELY):
- **Verbose explanations in config files** → VIOLATION: Write only essential value
- **Multiple lines explaining obvious behavior** → VIOLATION: One line maximum
- **"IMPORTANT:", "NOTE:", excessive formatting** → VIOLATION: Direct statement only

### 🟢 COMMENT ACTIONS ONLY:
- **Write as much words as needed to bring value to a professional**
- **State purpose, not process**
- **Essential information only**

## 💀 COMMENT ENFORCEMENT 💀

**WRONG:**
```typescript
// IMPORTANT: No webServer auto-start to ensure tests fail when services unavailable
// Tests should fail fast if frontend/backend not manually started  
// This prevents false passing tests when system is actually broken
```

**RIGHT:**
```typescript
/* No webServer auto-start - tests fail when services unavailable */
```

**NO EXCEPTIONS. NO NEGOTIATIONS. NO SURRENDER.**
