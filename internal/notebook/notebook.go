// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package notebook

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/symbolicsoft/enclave/v2/internal/ciphers"
	"github.com/symbolicsoft/enclave/v2/internal/client"
	"github.com/symbolicsoft/enclave/v2/internal/config"
	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
	"google.golang.org/protobuf/proto"
)

const NOTEBOOK_PAGE_BYTES_MAX = 64 * 1024
const NOTEBOOK_BYTES_MAX = NOTEBOOK_PAGE_BYTES_MAX * 128

func Create() *enclaveProto.Notebook {
	nb := &enclaveProto.Notebook{
		Pages: []*enclaveProto.Page{},
	}
	nb.Pages = append(nb.Pages, &enclaveProto.Page{
		Body:    "First page\n\nThis is an initial notebook page.",
		ModDate: time.Now().Unix(),
	})
	nb.Pages = append(nb.Pages, &enclaveProto.Page{
		Body:    "Second page\n\nThis is an example of a second page.",
		ModDate: time.Now().Unix(),
	})
	return nb
}

func Encrypt(sk ciphers.Subkey, nb *enclaveProto.Notebook) (ciphers.Ciphertext, error) {
	nbBytes, err := proto.Marshal(nb)
	if err != nil {
		return ciphers.Ciphertext{}, err
	}
	ct, err := ciphers.Encrypt(sk, nbBytes)
	if err != nil {
		return ciphers.Ciphertext{}, err
	}
	return ct, nil
}

func Decrypt(sk ciphers.Subkey, ct ciphers.Ciphertext) (*enclaveProto.Notebook, error) {
	pt, err := ciphers.Decrypt(sk, ct)
	if err != nil {
		return &enclaveProto.Notebook{}, err
	}
	nb := &enclaveProto.Notebook{}
	err = proto.Unmarshal(pt, nb)
	if err != nil {
		return &enclaveProto.Notebook{}, err
	}
	return nb, nil
}

func RestoreFromConfig() ([2]ciphers.Subkey, *enclaveProto.Notebook, error) {
	configFile, err := config.Read()
	if err != nil {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
	}
	uskId, errId := hex.DecodeString(configFile[0])
	uskEd, errEd := hex.DecodeString(configFile[1])
	if (errId != nil) || (errEd != nil) {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, errors.New("could not decode config file")
	}
	nb, err := Restore([2]ciphers.Subkey{uskId, uskEd})
	if err != nil {
		return [2]ciphers.Subkey{}, &enclaveProto.Notebook{}, err
	}
	return [2]ciphers.Subkey{uskId, uskEd}, nb, nil
}

func Restore(subkeys [2]ciphers.Subkey) (*enclaveProto.Notebook, error) {
	enb, err := client.GetNotebook(subkeys[0])
	if err != nil {
		return &enclaveProto.Notebook{}, err
	}
	ct := ciphers.Ciphertext{
		Data:  enb.Data,
		Nonce: enb.Nonce,
	}
	nb, err := Decrypt(subkeys[1], ct)
	if err != nil {
		return &enclaveProto.Notebook{}, err
	}
	return nb, nil
}
