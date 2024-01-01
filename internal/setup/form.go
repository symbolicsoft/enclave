// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package setup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/symbolicsoft/enclave/v2/internal/ciphers"
	"github.com/symbolicsoft/enclave/v2/internal/client"
	"github.com/symbolicsoft/enclave/v2/internal/version"
	"github.com/symbolicsoft/enclave/v2/internal/words"
)

func formHeader() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#34beed")).Bold(true).
		SetString(fmt.Sprintf("⎡ Enclave %s ", version.VERSION_CLIENT))
	return style.Render()
}

func formRestore() bool {
	var restore bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(strings.Join([]string{
					"Restore existing notebook?",
				}, "\n")).
				Description(strings.Join([]string{
					"You will be asked to enter your passphrase.",
					"If you don't have one, create a new notebook.",
				}, "\n")).
				Value(&restore),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return restore
}

func formCheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		defer cancel()
		errChan <- client.PingPong()
	}()
	spinner.New().Type(spinner.Dots).Title("Checking connection...").Context(ctx).Run()
	return <-errChan
}

func formRetryConnection() bool {
	var retry bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Retry connection?").
				Description("Connection failed.").
				Value(&retry),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return retry
}

func formPassphrase() string {
	var passphrase string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Enter your Enclave passphrase.").
				Description(strings.Join([]string{
					"Your Enclave passphrase and a network connection",
					"are all you need to restore your Enclave notebook locally.",
				}, "\n")).
				Validate(func(str string) error {
					matched, _ := regexp.MatchString(`^[a-z ]*$`, str)
					if !matched {
						return errors.New("invalid passphrase")
					}
					return nil
				}).
				Value(&passphrase),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return passphrase
}

func formConfirmCreateNotebook() bool {
	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create new Enclave notebook?").
				Description(strings.Join([]string{
					"Enclave will set up a notebook encrypted with a new passphrase",
					"and synchronize it with the Enclave server.",
				}, "\n")).
				Value(&confirm),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return confirm
}

func formCreateNotebook(decoyFor ciphers.Subkey) (string, [2]ciphers.Subkey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	passphrase := ""
	subkeys := [2]ciphers.Subkey{}
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		defer cancel()
		var enb ciphers.Ciphertext
		var err error
		passphrase, err = words.GeneratePassphrase(ciphers.PASSPHRASE_WORDS)
		if err != nil {
			errChan <- err
			return
		}
		subkeys, enb, err = setupNewNotebook(passphrase)
		if err != nil {
			errChan <- err
			return
		}
		client.PutNotebook(subkeys[0], decoyFor, enb)
	}()
	if len(decoyFor) > 0 {
		spinner.New().Type(spinner.Dots).Title("Creating decoy notebook...").Context(ctx).Run()
	} else {
		spinner.New().Type(spinner.Dots).Title("Creating notebook...").Context(ctx).Run()
	}
	return passphrase, subkeys, <-errChan
}

func formSetupDecoy() bool {
	var decoy bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Set up a decoy notebook?").
				Description(strings.Join([]string{
					"Decoy notebooks are accessible with their own, different passphrase.",
					"Accessing them more than once permanently deletes your \"real\" notebook,",
					"which you just created earlier.",
				}, "\n")).
				Value(&decoy),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return decoy
}

func formShowPassphrase(passphrase string, isDecoy bool) bool {
	var ready bool
	titleString := "Passphrase generated:"
	bodyString := strings.Join([]string{
		"Your passphrase is how you access your Enclave notebook from anywhere.",
		"Note your passphrase down somewhere safe before proceeding.",
		"",
		"Are you ready to proceed?",
	}, "\n")
	if isDecoy {
		titleString = "Decoy passphrase generated:"
		bodyString = strings.Join([]string{
			"This is your decoy passphrase. Loading the notebook associated with it",
			"more than once will delete the notebook you created earlier.",
			"",
			"Suggested usage:",
			"1. Load the decoy notebook once to input some decoy data.",
			"2. The next time the decoy notebook is loaded,",
			"   your non-decoy notebook will be deleted automatically.",
			"",
			"Remember that you can only load your decoy notebook once without",
			"the Enclave server deleting its associated non-decoy notebook.",
			"",
			"Are you ready to proceed?",
		}, "\n")
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(strings.Join([]string{titleString, passphrase}, "\n")).
				Description(bodyString).
				Value(&ready),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return ready
}

func formStoreKeysLocally() bool {
	var storeKeys bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Store access keys locally?").
				Description(strings.Join([]string{
					"You can choose to store your notebook access keys locally.",
					"This will launch your notebook immediately when opening Enclave",
					"without you having to type in your passphrase every time.",
					"",
					"However, it renders your keys vulnerable in case your computer is stolen.",
				}, "\n")).
				Value(&storeKeys),
		),
	).WithTheme(huh.ThemeBase16())
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	return storeKeys
}
