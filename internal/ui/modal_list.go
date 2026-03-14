// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

type ModalList struct {
	*tview.Box

	// The list embedded in the modal's frame.
	list *tview.List

	// The frame embedded in the modal.
	frame *tview.Frame

	// The optional callback for when the user clicked one of the items. It
	// receives the index of the clicked item and the item's text.
	done func(int, string)
}

func NewModalList(title string, list *tview.List) *ModalList {
	m := &ModalList{Box: tview.NewBox()}

	m.list = list
	m.list.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.list.SetSelectedFunc(func(i int, main string, _ string, _ rune) {
		if m.done != nil {
			m.done(i, main)
		}
	})
	m.list.SetDoneFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})

	m.frame = tview.NewFrame(m.list).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	m.frame.SetTitle(title)
	m.frame.SetTitleColor(tcell.ColorAqua)

	return m
}

// Draw draws this primitive onto the screen.
func (m *ModalList) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	width := 0
	for i := range m.list.GetItemCount() {
		main, secondary := m.list.GetItemText(i)
		width = max(width, len(main)+len(secondary)+2)
	}

	// Also consider the title length
	titleLen := len(m.frame.GetTitle()) + 4
	width = max(width, titleLen)

	// Set minimum width to ensure readability
	const minWidth = 70
	width = max(width, minWidth)

	screenWidth, screenHeight := screen.Size()

	// Ensure width doesn't exceed 80% of screen width
	maxWidth := (screenWidth * 4) / 5
	if width > maxWidth {
		width = maxWidth
	}

	// Set the modal's position and size.
	// Height calculation: items + frame borders(2) + frame padding(2) + header(1) + extra space(1) = items + 6
	height := m.list.GetItemCount() + 6
	// Ensure minimum height of 8 for proper single-item rendering with all borders and padding
	const minHeight = 8
	if height < minHeight {
		height = minHeight
	}
	width += 2
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

func (m *ModalList) SetDoneFunc(handler func(int, string)) *ModalList {
	m.done = handler
	return m
}

// Focus is called when this primitive receives focus.
func (m *ModalList) Focus(delegate func(p tview.Primitive)) {
	delegate(m.list)
}

// HasFocus returns whether this primitive has focus.
func (m *ModalList) HasFocus() bool {
	return m.list.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (m *ModalList) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.list.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *ModalList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}
