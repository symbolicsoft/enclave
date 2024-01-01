// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package setup

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/symbolicsoft/enclave/v2/internal/ciphers"
	"github.com/symbolicsoft/enclave/v2/internal/client"
	"github.com/symbolicsoft/enclave/v2/internal/config"
	"github.com/symbolicsoft/enclave/v2/internal/notebook"
	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
	"github.com/symbolicsoft/enclave/v2/internal/util"
)

func Setup() ([2]ciphers.Subkey, *enclaveProto.Notebook, error) {
	util.ClearManually()
	fmt.Println(formHeader())
	err := formCheckConnection()
	if err != nil {
		if formRetryConnection() {
			return Setup()
		} else {
			return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
		}
	}
	if formConfirmCreateNotebook() {
		passphrase, subkeys, err := formCreateNotebook(ciphers.Subkey{})
		if err != nil {
			return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
		}
		for !formShowPassphrase(passphrase, false) {
		}
		if formStoreKeysLocally() {
			err = config.Write([]string{
				hex.EncodeToString(subkeys[0]),
				hex.EncodeToString(subkeys[1]),
			})
			if err != nil {
				return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
			}
		}
		if formSetupDecoy() {
			decoyPassphrase, _, err := formCreateNotebook(subkeys[0])
			if err != nil {
				return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
			}
			for !formShowPassphrase(decoyPassphrase, true) {
			}
		}
		return subkeys, notebook.Create(), nil
	} else if formRestore() {
		passphrase := formPassphrase()
		subkeys, nb, err := setupGetNotebook(passphrase)
		if err != nil {
			return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
		}
		if formStoreKeysLocally() {
			err = config.Write([]string{
				hex.EncodeToString(subkeys[0]),
				hex.EncodeToString(subkeys[1]),
			})
			if err != nil {
				return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
			}
		}
		return subkeys, nb, nil
	}
	return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, errors.New("no notebook loaded")
}

func setupNewNotebook(passphrase string) ([2]ciphers.Subkey, ciphers.Ciphertext, error) {
	userSecret, err := ciphers.DeriveKey(passphrase)
	if err != nil {
		return [2]ciphers.Subkey{}, ciphers.Ciphertext{}, err
	}
	subkeys, err := ciphers.DeriveSubkeys(userSecret)
	if err != nil {
		return [2]ciphers.Subkey{}, ciphers.Ciphertext{}, err
	}
	nb := notebook.Create()
	enb, err := notebook.Encrypt(subkeys[1], nb)
	if err != nil {
		return [2]ciphers.Subkey{}, ciphers.Ciphertext{}, err
	}
	return subkeys, enb, err
}

func setupGetNotebook(passphrase string) ([2]ciphers.Subkey, *enclaveProto.Notebook, error) {
	userSecret, err := ciphers.DeriveKey(passphrase)
	if err != nil {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
	}
	subkeys, err := ciphers.DeriveSubkeys(userSecret)
	if err != nil {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
	}
	enb, err := client.GetNotebook(subkeys[0])
	if err != nil {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
	}
	ct := ciphers.Ciphertext{
		Data:  enb.Data,
		Nonce: enb.Nonce,
	}
	nb, err := notebook.Decrypt(subkeys[1], ct)
	if err != nil {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
	}
	return subkeys, nb, err
}
