// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/symbolicsoft/enclave/v2/internal/notebook"
)

type EditorModel struct {
	textarea textarea.Model
}

func (em EditorModel) Construct() EditorModel {
	ti := textarea.New()
	ti.Placeholder = "Ready..."
	ti.CharLimit = notebook.NOTEBOOK_PAGE_BYTES_MAX
	ti.KeyMap = textarea.KeyMap{
		CharacterForward:   key.NewBinding(key.WithKeys("right")),
		CharacterBackward:  key.NewBinding(key.WithKeys("left")),
		LineNext:           key.NewBinding(key.WithKeys("down")),
		LinePrevious:       key.NewBinding(key.WithKeys("up")),
		DeleteWordBackward: key.NewBinding(),
		DeleteAfterCursor: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "delete after cursor"),
		),
		DeleteBeforeCursor: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "delete before cursor"),
		),
		InsertNewline:              key.NewBinding(key.WithKeys("enter")),
		DeleteCharacterBackward:    key.NewBinding(key.WithKeys("backspace")),
		LineStart:                  key.NewBinding(key.WithKeys("home")),
		LineEnd:                    key.NewBinding(key.WithKeys("end")),
		Paste:                      key.NewBinding(key.WithKeys("ctrl+v")),
		InputBegin:                 key.NewBinding(key.WithKeys("alt+<")),
		InputEnd:                   key.NewBinding(key.WithKeys("alt+>")),
		DeleteWordForward:          key.NewBinding(),
		WordForward:                key.NewBinding(),
		WordBackward:               key.NewBinding(),
		DeleteCharacterForward:     key.NewBinding(),
		CapitalizeWordForward:      key.NewBinding(),
		LowercaseWordForward:       key.NewBinding(),
		UppercaseWordForward:       key.NewBinding(),
		TransposeCharacterBackward: key.NewBinding(),
	}
	ti.Focus()
	return EditorModel{
		textarea: ti,
	}
}

func (em EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (em EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	em.textarea, cmd = em.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return em, tea.Batch(cmds...)
}

func (em EditorModel) View() string {
	return em.textarea.View()
}
