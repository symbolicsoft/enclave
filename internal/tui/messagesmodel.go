// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MessageType uint

const (
	MessageInfo = iota
	MessageOK   = iota
	MessageErr  = iota
)

type MessagesModel struct {
	Messages string
	Width    int
	Height   int
}

func (mm MessagesModel) Construct() MessagesModel {
	return MessagesModel{
		Messages: "",
		Width:    10,
		Height:   2,
	}
}

func (mm MessagesModel) Init() tea.Cmd {
	return nil
}

func (mm MessagesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	return mm, tea.Batch(cmds...)
}

func (mm MessagesModel) View() string {
	return mm.Messages
}

func (mm *MessagesModel) SetMessage(msgType MessageType, message string) {
	var style lipgloss.Style
	switch msgType {
	case MessageInfo:
		style = lipgloss.NewStyle().Background(lipgloss.Color("#0000FF")).Bold(true).SetString(" INFO ")
	case MessageOK:
		style = lipgloss.NewStyle().Background(lipgloss.Color("#00FF00")).Bold(true).SetString("  OK  ")
	default:
		style = lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Bold(true).SetString("ERROR ")
	}
	mm.Messages = fmt.Sprintf("%s %s", style.Render(), message)
}

func (mm *MessagesModel) ClearMessage() {
	mm.Messages = ""
}
