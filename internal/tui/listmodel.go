// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
	"github.com/symbolicsoft/enclave/v2/internal/version"
)

type ListItem struct {
	page *enclaveProto.Page
}

func (li ListItem) Title() string {
	title := strings.Split(li.page.GetBody(), "\n")[0]
	if len(title) > 32 {
		title = title[:32] + "…"
	}
	return title
}

func (li ListItem) Description() string {
	return time.Unix(li.page.ModDate, 0).Format("Jan. 2, 2006 • 3:04PM")
}

func (li ListItem) FilterValue() string {
	return li.page.Body
}

type ListModel struct {
	list list.Model
}

func (lm ListModel) Construct(nb *enclaveProto.Notebook) ListModel {
	listItems := []list.Item{}
	for _, page := range nb.Pages {
		listItems = append(listItems, ListItem{page})
	}
	lm = ListModel{list: list.New(listItems, list.NewDefaultDelegate(), 0, 0)}
	lm.list.Title = fmt.Sprintf("Enclave %s", version.VERSION_CLIENT)
	lm.list.SetShowPagination(false)
	lm.list.SetShowStatusBar(true)
	lm.list.SetStatusBarItemName("page", "pages")
	// lm.list.SetShowFilter(true)
	// lm.list.SetFilteringEnabled(true)
	lm.list.SetShowHelp(true)
	lm.list.KeyMap = list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
	lm.list.Styles.TitleBar = lipgloss.NewStyle().Background(lipgloss.Color("#34beed"))
	return lm
}

func (lm ListModel) Init() tea.Cmd {
	return nil
}

func (lm ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	lm.list, cmd = lm.list.Update(msg)
	return lm, cmd
}

func (lm ListModel) View() string {
	return lm.list.View()
}
