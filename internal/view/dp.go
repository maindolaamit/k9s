// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const scaleDialogKey = "scale"

// Deploy represents a deployment view.
type Deploy struct {
	ResourceViewer
}

// NewDeploy returns a new deployment view.
func NewDeploy(gvr *client.GVR) ResourceViewer {
	var d Deploy
	d.ResourceViewer = NewPortForwardExtender(
		NewVulnerabilityExtender(
			NewRestartExtender(
				NewScaleExtender(
					NewImageExtender(
						NewOwnerExtender(
							NewLogsExtender(NewBrowser(gvr), d.logOptions),
						),
					),
				),
			),
		),
	)
	d.AddBindKeysFn(d.bindKeys)
	d.GetTable().SetEnterFn(d.showPods)

	return &d
}

func (d *Deploy) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyZ: ui.NewKeyAction("ReplicaSets", d.replicaSetsCmd, true),
		ui.KeyB: ui.NewKeyAction("Backup", d.backupCmd, false),
		ui.KeyU: ui.NewKeyAction("Restore", d.restoreCmd, false),
	})
}

func (d *Deploy) logOptions(prev bool) (*dao.LogOptions, error) {
	path := d.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}
	dp, err := d.getInstance(path)
	if err != nil {
		return nil, err
	}

	return podLogOptions(d.App(), path, prev, &dp.ObjectMeta, &dp.Spec.Template.Spec), nil
}

func (d *Deploy) replicaSetsCmd(evt *tcell.EventKey) *tcell.EventKey {
	dName := d.GetTable().GetSelectedItem()
	if dName == "" {
		return evt
	}
	dp, err := d.getInstance(dName)
	if err != nil {
		d.App().Flash().Err(err)
		return nil
	}
	showReplicasetsFromSelector(d.App(), dName, dp.Spec.Selector)
	return nil
}

func (d *Deploy) showPods(app *App, _ ui.Tabular, _ *client.GVR, fqn string) {
	dp, err := d.getInstance(fqn)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, fqn, dp.Spec.Selector)
}

func (d *Deploy) getInstance(fqn string) (*appsv1.Deployment, error) {
	var dp dao.Deployment
	dp.Init(d.App().factory, d.GVR())

	return dp.GetInstance(fqn)
}

// ----------------------------------------------------------------------------
// Helpers...

func showPodsFromSelector(app *App, path string, sel *metav1.LabelSelector) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, path, l, "")
}

func showReplicasetsFromSelector(app *App, path string, sel *metav1.LabelSelector) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showReplicasets(app, path, l, "")
}

func (d *Deploy) backupCmd(*tcell.EventKey) *tcell.EventKey {
	path := d.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	// Get deployment YAML
	dp, err := d.getInstance(path)
	if err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	// Get backup directory
	backupDir := filepath.Join(d.App().Config.K9s.GetScreenDumpDir(), "deployment-backups")
	if err := data.EnsureFullPath(backupDir, data.DefaultDirMod); err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	// Create backup file
	ns, name := client.Namespaced(path)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.yaml", ns, name, timestamp)
	backupPath := filepath.Join(backupDir, filename)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(dp)
	if err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	// Write to file
	if err := os.WriteFile(backupPath, yamlData, 0600); err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	d.App().Flash().Infof("Backup saved: %s", filename)
	return nil
}

func (d *Deploy) restoreCmd(*tcell.EventKey) *tcell.EventKey {
	path := d.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	ns, name := client.Namespaced(path)

	// Get backup directory
	backupDir := filepath.Join(d.App().Config.K9s.GetScreenDumpDir(), "deployment-backups")

	// Find backups for this deployment
	pattern := fmt.Sprintf("%s_%s_*.yaml", ns, name)
	matches, err := filepath.Glob(filepath.Join(backupDir, pattern))
	if err != nil {
		d.App().Flash().Err(err)
		return nil
	}

	if len(matches) == 0 {
		d.App().Flash().Warnf("No backups found for %s", path)
		return nil
	}

	// Sort by timestamp (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i] > matches[j]
	})

	// Show restore menu
	d.Stop()
	defer d.Start()
	d.showRestoreDialog(path, matches)

	return nil
}

func (d *Deploy) showRestoreDialog(path string, backups []string) {
	// Limit to 5 most recent backups
	if len(backups) > 5 {
		backups = backups[:5]
	}

	options := make([]string, len(backups))
	for i, backup := range backups {
		filename := filepath.Base(backup)
		// Extract timestamp from filename: namespace_name_20060102_150405.yaml
		parts := strings.Split(filename, "_")
		if len(parts) >= 4 {
			dateStr := parts[len(parts)-2]
			timeStr := strings.TrimSuffix(parts[len(parts)-1], ".yaml")

			// Parse and format timestamp
			timestamp := fmt.Sprintf("%s_%s", dateStr, timeStr)
			if t, err := time.Parse("20060102_150405", timestamp); err == nil {
				options[i] = fmt.Sprintf("[%d] %s", i, t.Format("2006-01-02 15:04:05"))
			} else {
				options[i] = fmt.Sprintf("[%d] %s", i, filename)
			}
		} else {
			options[i] = fmt.Sprintf("[%d] %s", i, filename)
		}
	}

	restoreAction := func(index int) {
		if index < 0 || index >= len(backups) {
			return
		}

		backupFile := backups[index]

		// Read backup file
		yamlData, err := os.ReadFile(backupFile)
		if err != nil {
			d.App().Flash().Err(err)
			return
		}

		// Parse deployment
		var dp appsv1.Deployment
		if err := yaml.Unmarshal(yamlData, &dp); err != nil {
			d.App().Flash().Err(err)
			return
		}

		// Get current deployment and update its spec
		ctx, cancel := context.WithTimeout(context.Background(), d.App().Conn().Config().CallTimeout())
		defer cancel()

		current, err := d.getInstance(path)
		if err != nil {
			d.App().Flash().Err(err)
			return
		}

		// Update spec from backup
		current.Spec = dp.Spec

		// Apply via Kubernetes API
		dial, err := d.App().Conn().Dial()
		if err != nil {
			d.App().Flash().Err(err)
			return
		}

		ns, _ := client.Namespaced(path)
		if _, err := dial.AppsV1().Deployments(ns).Update(ctx, current, metav1.UpdateOptions{}); err != nil {
			d.App().Flash().Err(err)
			return
		}

		d.App().Flash().Infof("Restored deployment %s from backup", path)
	}

	styles := d.App().Styles.Dialog()
	list := tview.NewList().ShowSecondaryText(false)
	list.SetSelectedTextColor(styles.ButtonFocusFgColor.Color())
	list.SetSelectedBackgroundColor(styles.ButtonFocusBgColor.Color())
	list.SetMainTextColor(styles.FgColor.Color())

	for _, option := range options {
		list.AddItem(option, "", 0, nil)
	}

	// Ensure first item is selected
	if len(options) > 0 {
		list.SetCurrentItem(0)
	}

	// Add keyboard shortcuts
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle 0-4 for direct selection
		if event.Rune() >= '0' && event.Rune() <= '4' {
			index := int(event.Rune() - '0')
			if index < len(backups) {
				d.dismissRestoreDialog()
				restoreAction(index)
			}
			return nil
		}
		switch event.Rune() {
		case 'q', 'Q':
			d.dismissRestoreDialog()
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			d.dismissRestoreDialog()
			return nil
		}
		return event
	})

	modal := ui.NewModalList(fmt.Sprintf("<Restore %s>", path), list)
	modal.SetDoneFunc(func(i int, _ string) {
		d.dismissRestoreDialog()
		if i >= 0 && i < len(backups) {
			restoreAction(i)
		}
	})

	d.App().Content.AddPage("restore", modal, true, true)
	d.App().Content.ShowPage("restore")
}

func (d *Deploy) dismissRestoreDialog() {
	d.App().Content.RemovePage("restore")
}
