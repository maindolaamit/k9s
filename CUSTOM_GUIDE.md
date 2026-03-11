# K9s Custom Fork - Complete Guide

## Quick Start

```bash
# 1. Build your custom k9s
./build.sh

# 2. Test it (safe - won't replace current k9s)
cp execs/k9s /usr/local/bin/k9s-custom
k9s-custom version

# 3. Make customizations (see examples below)
# 4. Rebuild and test
```

---

## Why This Fork?

Standard k9s config (`~/.config/k9s/hotkeys.yaml`) **cannot override** built-in shortcuts like `:`, `?`, `Ctrl-C`, etc.

This fork lets you:
- Override ANY built-in shortcut
- Change default settings (refresh rate, default view, etc.)
- Add deeply integrated custom features
- Customize for your project needs

Your existing plugins will continue working.

---

## Key Files to Modify

| What to Change | File | Line Hint |
|----------------|------|-----------|
| App shortcuts (help, quit, etc.) | `internal/view/app.go` | ~254 (bindKeys) |
| Default config | `internal/config/k9s.go` | ~45 (newK9s) |
| UI shortcuts | `internal/ui/app.go` | ~150 (bindKeys) |

---

## Common Customizations

### 1. Add Custom Namespace Shortcut (Ctrl-N)

**File:** `internal/view/app.go` (line 254, in `bindKeys()` function)

**Add after line 262:**
```go
tcell.KeyCtrlN: ui.NewSharedKeyAction("Jump to app00739", a.jumpToApp00739, false),
```

**Add handler function after bindKeys():**
```go
func (a *App) jumpToApp00739(evt *tcell.EventKey) *tcell.EventKey {
    if err := a.Config.SetActiveNamespace("app00739"); err != nil {
        a.Flash().Errf("Failed: %s", err)
        return nil
    }
    a.gotoResource("pods", "", false, false)
    a.Flash().Info("Switched to app00739")
    return nil
}
```

### 2. Change Quit Key (Ctrl-C → Ctrl-Q)

**File:** `internal/view/app.go` (line 264)

**Change:**
```go
tcell.KeyCtrlC: ui.NewKeyAction("Quit", a.quitCmd, false),
```

**To:**
```go
tcell.KeyCtrlQ: ui.NewKeyAction("Quit", a.quitCmd, false),
```

### 3. Change Help Key (? → F1)

**File:** `internal/view/app.go` (line 258)

**Change:**
```go
ui.KeyHelp: ui.NewSharedKeyAction("Help", a.helpCmd, false),
```

**To:**
```go
tcell.KeyF1: ui.NewSharedKeyAction("Help", a.helpCmd, false),
```

### 4. Faster Refresh Rate (2s → 1s)

**File:** `internal/config/k9s.go`

**Find `newK9s()` function and change:**
```go
func newK9s() *K9s {
    return &K9s{
        RefreshRate: 1,  // Changed from 2 to 1 second
        // ...
    }
}
```

### 5. Default to Pods View

**File:** `internal/config/k9s.go`

**In `newK9s()` function, add:**
```go
DefaultView: "pods",  // Always start with pods
```

### 6. Add Multiple Custom Shortcuts

**File:** `internal/view/app.go`

**Complete example with multiple shortcuts:**
```go
func (a *App) bindKeys() {
    a.AddActions(ui.NewKeyActionsFromMap(ui.KeyMap{
        tcell.KeyCtrlE:     ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
        tcell.KeyCtrlG:     ui.NewSharedKeyAction("ToggleCrumbs", a.toggleCrumbsCmd, false),
        tcell.KeyF1:        ui.NewSharedKeyAction("Help", a.helpCmd, false), // Changed from ?
        ui.KeyLeftBracket:  ui.NewSharedKeyAction("Go Back", a.previousCommand, false),
        ui.KeyRightBracket: ui.NewSharedKeyAction("Go Forward", a.nextCommand, false),
        ui.KeyDash:         ui.NewSharedKeyAction("Last View", a.lastCommand, false),
        tcell.KeyCtrlA:     ui.NewSharedKeyAction("Aliases", a.aliasCmd, false),

        // YOUR CUSTOM SHORTCUTS
        tcell.KeyCtrlN:     ui.NewSharedKeyAction("Jump to app00739", a.jumpToApp00739, false),
        tcell.KeyCtrlK:     ui.NewSharedKeyAction("Quick Pods", a.quickPods, false),
        tcell.KeyF2:        ui.NewSharedKeyAction("Quick Deployments", a.quickDeploy, false),

        tcell.KeyEnter:     ui.NewKeyAction("Goto", a.gotoCmd, false),
        tcell.KeyCtrlQ:     ui.NewKeyAction("Quit", a.quitCmd, false), // Changed from Ctrl-C
    }))
}

// Handler functions
func (a *App) jumpToApp00739(evt *tcell.EventKey) *tcell.EventKey {
    a.Config.SetActiveNamespace("app00739")
    a.gotoResource("pods", "", false, false)
    a.Flash().Info("Switched to app00739")
    return nil
}

func (a *App) quickPods(evt *tcell.EventKey) *tcell.EventKey {
    a.gotoResource("pods", "", false, false)
    return nil
}

func (a *App) quickDeploy(evt *tcell.EventKey) *tcell.EventKey {
    a.gotoResource("deployments", "", false, false)
    return nil
}
```

---

## Available Key Codes

```go
// Control keys
tcell.KeyCtrlA through tcell.KeyCtrlZ

// Function keys
tcell.KeyF1 through tcell.KeyF12

// Special keys
tcell.KeyEnter, tcell.KeyEsc, tcell.KeyTab
tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight
tcell.KeyBackspace, tcell.KeyDelete
tcell.KeyHome, tcell.KeyEnd, tcell.KeyPgUp, tcell.KeyPgDn

// K9s custom keys (from internal/ui/key.go)
ui.KeyHelp (?)
ui.KeySlash (/)
ui.KeySpace
ui.KeyA through ui.KeyZ
ui.Key0 through ui.Key9
```

---

## Build and Deploy

### Build
```bash
./build.sh
```

### Test (Recommended)
```bash
# Install as separate binary
cp execs/k9s /usr/local/bin/k9s-custom
k9s-custom

# Your original k9s is untouched
```

### Deploy (After Testing)
```bash
# Backup original
cp /opt/homebrew/bin/k9s /opt/homebrew/bin/k9s.bak

# Replace with custom version
cp execs/k9s /opt/homebrew/bin/k9s
```

---

## Workflow

```
1. Edit file (e.g., internal/view/app.go)
   ↓
2. ./build.sh
   ↓
3. Test with k9s-custom
   ↓
4. Document in CUSTOMIZATIONS.txt
   ↓
5. git commit -m "Add feature"
   ↓
6. Deploy when ready
```

---

## Update from Upstream

```bash
# One-time setup
git remote add upstream https://github.com/derailed/k9s.git

# Update
git fetch upstream
git merge upstream/master
# Resolve any conflicts
./build.sh
```

---

## Troubleshooting

### Build fails
```bash
go mod tidy
./build.sh
```

### Changes don't appear
```bash
# Check which binary you're running
which k9s

# Clear cache
rm -rf ~/.config/k9s/.cache/

# Rebuild from scratch
rm -rf execs/ && ./build.sh
```

### Merge conflicts
```bash
git stash                    # Save changes
git merge upstream/master    # Merge
# Resolve conflicts
git stash pop               # Reapply changes
./build.sh
```

---

## Track Your Changes

Use `CUSTOMIZATIONS.txt` to track what you've modified:

```
# My K9s Customizations

## Implemented
- [x] Ctrl-N: Jump to app00739 (internal/view/app.go:262)
- [x] F1: Help instead of ? (internal/view/app.go:258)
- [x] Refresh rate: 1s (internal/config/k9s.go:45)

## Planned
- [ ] Custom dashboard view
- [ ] Auto-backup before dangerous ops
```

---

## Your Specific Setup

**Namespace:** app00739
**Sudo:** `--as app00739-sudo`
**Plugins:** All your plugins in `~/.config/k9s/plugins.yaml` will continue working

Suggested customizations for your workflow:
1. Quick namespace shortcuts (Ctrl-N for app00739)
2. Default to your namespace
3. Faster refresh rate
4. Custom shortcuts for deployments view

---

## Implemented Custom Keybindings

### Navigation (Vim-style)
- **Ctrl-U**: Page Up (half-page scroll up)
- **Ctrl-D**: Page Down (half-page scroll down)
  - *Changed from:* Ctrl-B (page up), Ctrl-F (page down)
  - *Files modified:* `internal/ui/select_table.go`, `internal/view/table.go`

### Clear Filter
- **Ctrl-L**: Clear filter/command
- **Ctrl-Q**: Clear filter (backup)
  - *Changed from:* Ctrl-U
  - *Files modified:* `internal/ui/app.go`

### Delete Resource
- **Shift-D**: Delete resource (normal delete)
  - *Changed from:* Ctrl-D
  - *Files modified:* Multiple view files (browser.go, workload.go, xray.go, etc.)

### Delete AsUser (Privileged Delete)
- **Ctrl-K**: Delete resource with impersonation
  - *Requires:* `asUser` configured in `~/.config/k9s/config.yaml`
  - *Example config:*
    ```yaml
    k9s:
      asUser: app00739-sudo
    ```
  - *Files modified:* `internal/config/k9s.go`, view files

### Configuration
Add the following to your `~/.config/k9s/config.yaml`:
```yaml
k9s:
  asUser: app00739-sudo  # User for Ctrl-K privileged delete
```

**Note:** The AsUser delete (Ctrl-K) feature currently shows a warning that it's not fully implemented. Full impersonation support for delete operations is planned for future development.

---

## Tips

- Test with `k9s-custom` before replacing your main binary
- Start with one change at a time
- Document changes in CUSTOMIZATIONS.txt
- Commit frequently
- Keep a backup of your original k9s

---

## Resources

- Original k9s: https://github.com/derailed/k9s
- tcell (keyboard): https://github.com/gdamore/tcell
- This fork location: `/Users/l9w9/github/k9s-custom`
- Branch: `custom-keybindings`

---

## Next Steps

1. Build: `./build.sh`
2. Test: `cp execs/k9s /usr/local/bin/k9s-custom && k9s-custom`
3. Customize: Edit `internal/view/app.go` or `internal/config/k9s.go`
4. Rebuild and test
5. Deploy when ready

Happy customizing!
