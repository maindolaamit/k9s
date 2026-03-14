// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/labels"
)

// Secret presents a secret viewer.
type Secret struct {
	ResourceViewer
}

// NewSecret returns a new viewer.
func NewSecret(gvr *client.GVR) ResourceViewer {
	s := Secret{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *Secret) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyX:      ui.NewKeyAction("Decode", s.decodeCmd, true),
		ui.KeyShiftX: ui.NewKeyAction("Export", s.exportCmd, true),
		ui.KeyU:      ui.NewKeyAction("UsedBy", s.refCmd, true),
	})
}

func (s *Secret) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), client.SecGVR)
}

func (s *Secret) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	o, err := s.App().factory.Get(s.GVR(), path, true, labels.Everything())
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	mm, err := dao.ExtractSecrets(o)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	raw, err := data.WriteYAML(mm)
	if err != nil {
		s.App().Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	details := NewDetails(s.App(), "Secret Decoder", path, contentYAML, true).Update(string(raw))
	if err := s.App().inject(details, false); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func (s *Secret) exportCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the secret
	o, err := s.App().factory.Get(s.GVR(), path, true, labels.Everything())
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	// Decode the secret
	mm, err := dao.ExtractSecrets(o)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	// Convert to YAML
	raw, err := data.WriteYAML(mm)
	if err != nil {
		s.App().Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	// Get export directory
	exportDir := s.App().Config.K9s.GetScreenDumpDir()
	secretsDir := filepath.Join(exportDir, "secrets")
	if err := data.EnsureFullPath(secretsDir, data.DefaultDirMod); err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	// Create filename: namespace_secretname_timestamp.yaml
	ns, name := client.Namespaced(path)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.yaml", ns, name, timestamp)
	// Sanitize filename
	filename = strings.ReplaceAll(filename, "/", "_")
	exportPath := filepath.Join(secretsDir, filename)

	// Write to file
	if err := os.WriteFile(exportPath, raw, 0600); err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	s.App().Flash().Infof("Secret exported: %s", filename)
	return nil
}
