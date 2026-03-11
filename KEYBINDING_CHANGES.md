# Keybinding Changes Summary

## Overview
This document summarizes all the custom keybinding changes implemented in this k9s fork.

## Changes Implemented

### 1. Vim-Style Page Navigation ✅
**Changed:**
- `Ctrl-U` → Page Up (half-page scroll up)
- `Ctrl-D` → Page Down (half-page scroll down)

**Previous:**
- `Ctrl-B` → Page Up
- `Ctrl-F` → Page Down

**Files Modified:**
- `internal/ui/select_table.go` - Added PageUp() and PageDown() methods
- `internal/view/table.go` - Added keyboard handler for Ctrl-U/D
- `internal/view/help.go` - Updated navigation help text

---

### 2. Clear Filter Keybinding ✅
**Changed:**
- `Ctrl-L` → Clear Filter/Command (primary)
- `Ctrl-Q` → Clear Filter (backup)

**Previous:**
- `Ctrl-U` → Clear Filter
- `Ctrl-Q` → Clear Filter (backup)

**Files Modified:**
- `internal/ui/app.go` - Changed binding from Ctrl-U to Ctrl-L
- `internal/view/help.go` - Updated general help text

**Note:** Ctrl-L may be overridden in ReplicaSet view where it's used for Rollback.

---

### 3. Delete Resource Keybinding ✅
**Changed:**
- `Shift-D` → Delete resource

**Previous:**
- `Ctrl-D` → Delete resource

**Files Modified:**
- `internal/view/browser.go` - Changed delete binding
- `internal/view/workload.go` - Changed delete binding
- `internal/view/xray.go` - Changed delete binding
- `internal/view/context.go` - Changed delete binding
- `internal/view/pf.go` - Changed delete binding
- `internal/view/user.go` - Updated deleted keys list
- `internal/view/dir.go` - Updated deleted keys list
- `internal/view/event.go` - Updated deleted keys list
- `internal/view/helm_history.go` - Updated deleted keys list

---

### 4. AsUser Configuration ✅
**Added:**
New configuration field for user impersonation.

**Configuration:**
Add to `~/.config/k9s/config.yaml`:
```yaml
k9s:
  asUser: app00739-sudo  # Your privileged user
```

**Files Modified:**
- `internal/config/k9s.go` - Added AsUser string field

---

### 5. AsUser Delete (Privileged Delete) ⚠️
**Added:**
- `Ctrl-K` → Delete resource as configured AsUser

**Status:** Keybinding implemented, full functionality pending.

**Current Behavior:**
- Shows warning that AsUser delete is not yet fully implemented
- Displays configured AsUser from config
- Prompts user to use Shift-D for regular delete

**Files Modified:**
- `internal/view/browser.go` - Added asUserDeleteCmd method and Ctrl-K binding
- `internal/view/workload.go` - Added Ctrl-K binding
- `internal/view/xray.go` - Added asUserDeleteCmd method and Ctrl-K binding

**Future Work:**
- Implement Kubernetes client impersonation for delete operations
- Support temporary client configuration with --as flag
- Or: Shell out to kubectl with --as flag for delete operations

---

## Quick Reference

| Action | Old Keybinding | New Keybinding |
|--------|----------------|----------------|
| Page Up | Ctrl-B | **Ctrl-U** |
| Page Down | Ctrl-F | **Ctrl-D** |
| Clear Filter | Ctrl-U | **Ctrl-L** |
| Delete Resource | Ctrl-D | **Shift-D** |
| Delete AsUser | N/A | **Ctrl-K** (new) |

---

## Testing

After building, test the following:

1. **Page Navigation:**
   ```
   - Open any resource view (e.g., pods)
   - Press Ctrl-U to scroll up
   - Press Ctrl-D to scroll down
   ```

2. **Clear Filter:**
   ```
   - Apply a filter with /
   - Press Ctrl-L to clear
   - Or press Ctrl-Q to clear (backup)
   ```

3. **Delete Resource:**
   ```
   - Select a resource
   - Press Shift-D to delete (confirmation dialog)
   ```

4. **AsUser Configuration:**
   ```
   - Add asUser to ~/.config/k9s/config.yaml
   - Select a resource
   - Press Ctrl-K (shows warning message)
   ```

---

## Build and Deploy

```bash
# Build
./build.sh

# Test (safe - won't replace current k9s)
cp execs/k9s /usr/local/bin/k9s-custom
k9s-custom version

# If all tests pass, deploy
cp execs/k9s /opt/homebrew/bin/k9s
# or
cp execs/k9s /usr/local/bin/k9s
```

---

## Rollback

If you need to restore original k9s:

```bash
# If you have a backup
cp /opt/homebrew/bin/k9s.bak /opt/homebrew/bin/k9s

# Or reinstall via Homebrew
brew reinstall derailed/k9s/k9s
```

---

## Notes

- All keybindings are hardcoded in the Go source
- Help text is dynamically generated from keybindings
- Changes require rebuild with `./build.sh`
- Configuration (asUser) can be changed in config.yaml without rebuild

---

Last Updated: 2026-03-10
Branch: custom-keybindings
