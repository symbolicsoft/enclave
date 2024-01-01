// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package tui

import (
	"bufio"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/symbolicsoft/enclave/v2/internal/ciphers"
	"github.com/symbolicsoft/enclave/v2/internal/client"
	"github.com/symbolicsoft/enclave/v2/internal/config"
	"github.com/symbolicsoft/enclave/v2/internal/notebook"
	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
	"github.com/symbolicsoft/enclave/v2/internal/setup"
)

var (
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	listStyleFocused = lipgloss.NewStyle().
				Align(lipgloss.Left, lipgloss.Center).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#34beed"))
	editorStyle = lipgloss.NewStyle().
			Align(lipgloss.Right, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	editorStyleFocused = lipgloss.NewStyle().
				Align(lipgloss.Right, lipgloss.Center).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#34beed"))
	messagesStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center)
)

type MainModel struct {
	list        ListModel
	editor      EditorModel
	messages    MessagesModel
	focusedView uint
	uskId       ciphers.Subkey
	uskEd       ciphers.Subkey
	notebook    *enclaveProto.Notebook
	pageIndex   int
}

func (mm MainModel) Construct(subkeys [2]ciphers.Subkey, nb *enclaveProto.Notebook) MainModel {
	if len(nb.Pages) == 0 {
		nb = notebook.Create()
	}
	mm = MainModel{
		list:        ListModel{}.Construct(nb),
		editor:      EditorModel{}.Construct(),
		messages:    MessagesModel{}.Construct(),
		focusedView: 0,
		uskId:       subkeys[0],
		uskEd:       subkeys[1],
		notebook:    nb,
		pageIndex:   0,
	}
	mm.editor.textarea.SetValue(mm.notebook.Pages[0].Body)
	return mm
}

func (mm MainModel) Init() tea.Cmd {
	return nil
}

func (mm MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	updateNotebook := false
	previousValue := ""
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch mm.focusedView {
		case 0:
			switch msg.String() {
			case "enter":
				mm.pageIndex = mm.list.list.Index()
				mm.editor.textarea.SetValue(mm.notebook.Pages[mm.pageIndex].Body)
				mm.focusedView = 1
				mm.editor.textarea.Focus()
			case "tab":
				mm.focusedView = 1
				mm.editor.textarea.Focus()
			case "ctrl+a":
				newPage := &enclaveProto.Page{
					Body:    "New page\n\n",
					ModDate: time.Now().Unix(),
				}
				mm.notebook.Pages = append([]*enclaveProto.Page{newPage}, mm.notebook.Pages...)
				mm.list.list.InsertItem(0, ListItem{newPage})
				mm.pageIndex = 0
				mm.editor.textarea.SetValue(mm.notebook.Pages[mm.pageIndex].Body)
				mm.focusedView = 1
				mm.editor.textarea.Focus()
				mm.messages.SetMessage(MessageInfo, "Page created.")
			case "ctrl+d":
				if len(mm.list.list.Items()) > 0 {
					listPageIndex := mm.list.list.Index()
					mm.list.list.RemoveItem(listPageIndex)
					mm.notebook.Pages = append(mm.notebook.Pages[:listPageIndex], mm.notebook.Pages[listPageIndex+1:]...)
					mm.messages.SetMessage(MessageInfo, "Page deleted.")
				}
			case "ctrl+s":
				mm.messages.SetMessage(MessageInfo, "Saving notebook...")
				mm.saveNotebook()
			case "ctrl+c":
				return mm, tea.Quit
			}
		default:
			switch msg.String() {
			case "tab":
				mm.focusedView = 0
				mm.editor.textarea.Blur()
			case "ctrl+s":
				mm.messages.SetMessage(MessageInfo, "Saving notebook...")
				mm.saveNotebook()
			case "ctrl+c":
				return mm, tea.Quit
			default:
				updateNotebook = true
				previousValue = mm.editor.textarea.Value()
			}
		}
		switch mm.focusedView {
		case 0:
			mmNew, cmd := mm.list.Update(msg)
			mm.list = mmNew.(ListModel)
			cmds = append(cmds, cmd)
		default:
			mmNew, cmd := mm.editor.Update(msg)
			mm.editor = mmNew.(EditorModel)
			cmds = append(cmds, cmd)
		}
		if updateNotebook && previousValue != mm.editor.textarea.Value() {
			mm.notebook.Pages[mm.pageIndex].Body = mm.editor.textarea.Value()
			mm.notebook.Pages[mm.pageIndex].ModDate = time.Now().Unix()
		}
	case tea.WindowSizeMsg:
		lR, lC := (30 * (msg.Width) / 100), (msg.Height - 3)
		eR, eC := (73 * (msg.Width) / 100), (msg.Height - 3)
		mm.list.list.SetSize(lR, lC)
		mm.list.list.SetWidth(lR)
		mm.list.list.SetHeight(lC)
		mm.editor.textarea.SetWidth(eR)
		mm.editor.textarea.SetHeight(eC)
		mm.messages.Width = (msg.Width - 3)
		mm.messages.Height = 1
	}
	return mm, tea.Batch(cmds...)
}

func (mm MainModel) View() string {
	var s string
	if mm.focusedView == 0 {
		s += lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Center,
				listStyleFocused.Render(mm.list.View()),
				editorStyle.Render(mm.editor.View()),
			),
			messagesStyle.Render(mm.messages.View()),
		)
	} else {
		s += lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Center,
				listStyle.Render(mm.list.View()),
				editorStyleFocused.Render(mm.editor.View()),
			),
			messagesStyle.Render(mm.messages.View()),
		)
	}
	return s
}

func (mm *MainModel) saveNotebook() {
	ct, err := notebook.Encrypt(mm.uskEd, mm.notebook)
	if err != nil {
		mm.messages.SetMessage(MessageErr, err.Error())
	} else {
		err = client.PutNotebook(mm.uskId, []byte{}, ct)
		if err != nil {
			mm.messages.SetMessage(MessageErr, err.Error())
		} else {
			mm.messages.SetMessage(MessageOK, "Notebook saved.")
		}
	}
}

func RunProgram() {
	if config.ConfigFileExists() != nil {
		subkeys, nb, err := setup.Setup()
		if err != nil {
			offerToRestart(err)
			return
		}
		mainModel := MainModel{}.Construct(subkeys, nb)
		runEditorTui(mainModel)
	} else {
		subkeys, nb, err := notebook.RestoreFromConfig()
		if err != nil {
			offerToRestart(err)
			return
		}
		mainModel := MainModel{}.Construct(subkeys, nb)
		runEditorTui(mainModel)
	}
}

func offerToRestart(err error) {
	fmt.Println(err)
	fmt.Print("Press 'Enter' to restart...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	RunProgram()
}

func runEditorTui(mainModel MainModel) {
	p := tea.NewProgram(mainModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
