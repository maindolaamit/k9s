// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/labels"
)

// ConfigMap represents a configmap viewer.
type ConfigMap struct {
	ResourceViewer
}

// NewConfigMap returns a new viewer.
func NewConfigMap(gvr *client.GVR) ResourceViewer {
	s := ConfigMap{
		ResourceViewer: NewOwnerExtender(
			NewBrowser(gvr),
		),
	}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *ConfigMap) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyX:      ui.NewKeyAction("Decode", s.decodeCmd, true),
		ui.KeyShiftX: ui.NewKeyAction("Export", s.exportCmd, true),
		ui.KeyU:      ui.NewKeyAction("UsedBy", s.refCmd, true),
	})
}

func (s *ConfigMap) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), client.CmGVR)
}

func (s *ConfigMap) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	o, err := s.App().factory.Get(s.GVR(), path, true, labels.Everything())
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	mm, err := dao.ExtractConfigMap(o)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	raw, err := data.WriteYAML(mm)
	if err != nil {
		s.App().Flash().Errf("Error extracting configmap %s", err)
		return nil
	}

	details := NewDetails(s.App(), "ConfigMap Viewer", path, contentYAML, true).Update(string(raw))
	if err := s.App().inject(details, false); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func (s *ConfigMap) exportCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the configmap
	o, err := s.App().factory.Get(s.GVR(), path, true, labels.Everything())
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	// Extract the configmap data
	mm, err := dao.ExtractConfigMap(o)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	// Convert to YAML
	raw, err := data.WriteYAML(mm)
	if err != nil {
		s.App().Flash().Errf("Error extracting configmap %s", err)
		return nil
	}

	// Get export directory
	exportDir := s.App().Config.K9s.GetScreenDumpDir()
	configmapsDir := filepath.Join(exportDir, "configmaps")
	if err := data.EnsureFullPath(configmapsDir, data.DefaultDirMod); err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	// Create filename: namespace_configmapname_timestamp.yaml
	ns, name := client.Namespaced(path)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.yaml", ns, name, timestamp)
	// Sanitize filename
	filename = strings.ReplaceAll(filename, "/", "_")
	exportPath := filepath.Join(configmapsDir, filename)

	// Write to file
	if err := os.WriteFile(exportPath, raw, 0600); err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	s.App().Flash().Infof("ConfigMap exported: %s", filename)
	return nil
}

func scanRefs(evt *tcell.EventKey, a *App, t *Table, gvr *client.GVR) *tcell.EventKey {
	path := t.GetSelectedItem()
	if path == "" {
		return evt
	}

	ctx := context.Background()
	refs, err := dao.ScanForRefs(refContext(gvr, path, true)(ctx), a.factory)
	if err != nil {
		a.Flash().Err(err)
		return nil
	}
	if len(refs) == 0 {
		a.Flash().Warnf("No references found at this time for %s::%s. Check again later!", gvr, path)
		return nil
	}
	a.Flash().Infof("Viewing references for %s::%s", gvr, path)
	view := NewReference(client.RefGVR)
	view.SetContextFn(refContext(gvr, path, false))
	if err := a.inject(view, false); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func refContext(gvr *client.GVR, path string, wait bool) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		ctx = context.WithValue(ctx, internal.KeyGVR, gvr)
		return context.WithValue(ctx, internal.KeyWait, wait)
	}
}
