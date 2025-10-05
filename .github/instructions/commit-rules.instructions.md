---
applyTo: '**'
---

# 🚨 COMMIT RULES PROTOCOL 🚨

## ⚡ COMMIT MESSAGE ENFORCEMENT ⚡

### 🔴 COMMIT RED FLAGS (STOP IMMEDIATELY):

- **Emojis in commit messages** → VIOLATION: Professional commits only
- **Multiple sentences** → VIOLATION: Single concise phrase only
- **Vague descriptions ("fix bug", "update code")** → VIOLATION: Specific action required
- **Excessive words (>8 words)** → VIOLATION: Blunt description only

### 🟢 COMMIT ACTIONS ONLY:

- **"Add [specific feature/file]"**
- **"Fix [specific issue]"**
- **"Remove [specific component]"**
- **"Update [specific functionality]"**

## 💀 COMMIT ENFORCEMENT 💀

**WRONG:**

```bash
git commit -m "✨ Added some new features and fixed various bugs in the application 🚀"
```

**RIGHT:**

```bash
git commit -m "Add Node.js app with local PineTS integration"
```

**NO EXCEPTIONS. NO NEGOTIATIONS. NO SURRENDER.**
