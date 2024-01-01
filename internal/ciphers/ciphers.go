// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package ciphers

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/symbolicsoft/enclave/v2/internal/words"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
)

const SCRYPT_N = 1048576
const SCRYPT_R = 8
const SCRYPT_P = 1
const SCRYPT_L = 32
const SCRYPT_SALT = "DTWdTA8L9VZG5J8p5dNaUmrQ"
const SUBKEY_L = 32
const PASSPHRASE_WORDS = 12

type Key []byte
type Subkey []byte

type Ciphertext struct {
	Data  []byte
	Nonce []byte
}

func GeneratePassphrase() (string, error) {
	return words.GeneratePassphrase(PASSPHRASE_WORDS)
}

func DeriveKey(passphrase string) (Key, error) {
	spaces := 0
	for _, char := range passphrase {
		if char == ' ' {
			spaces++
		}
	}
	if spaces < (PASSPHRASE_WORDS - 1) {
		return []byte{}, fmt.Errorf("passphrase must have at least %d words", PASSPHRASE_WORDS)
	}
	salt := []byte(SCRYPT_SALT)
	key, err := scrypt.Key([]byte(passphrase), salt, SCRYPT_N, SCRYPT_R, SCRYPT_P, SCRYPT_L)
	if err != nil {
		return nil, err
	}
	return key, err
}

func DeriveSubkeys(k Key) ([2]Subkey, error) {
	if len(k) != SCRYPT_L {
		return [2]Subkey{}, fmt.Errorf("derived key must be %d bytes", SUBKEY_L)
	}
	blake2x, err := blake2s.NewXOF(SUBKEY_L*2, k)
	if err != nil {
		return [2]Subkey{}, err
	}
	subkeys := [2]Subkey{make([]byte, SUBKEY_L), make([]byte, SUBKEY_L)}
	skn0, err := blake2x.Read(subkeys[0])
	if skn0 != SUBKEY_L {
		return [2]Subkey{}, errors.New("could not derive full subkey")
	}
	if err != nil {
		return [2]Subkey{}, err
	}
	skn1, err := blake2x.Read(subkeys[1])
	if skn1 != SUBKEY_L {
		return [2]Subkey{}, errors.New("could not derive full subkey")
	}
	if err != nil {
		return [2]Subkey{}, err
	}
	return subkeys, err
}

func Encrypt(sk Subkey, pt []byte) (Ciphertext, error) {
	if len(sk) != SUBKEY_L {
		return Ciphertext{}, fmt.Errorf("encryption key must be %d bytes", SUBKEY_L)
	}
	cipher, err := chacha20poly1305.NewX(sk)
	if err != nil {
		return Ciphertext{}, err
	}
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	n, err := rand.Read(nonce)
	if n != chacha20poly1305.NonceSizeX {
		return Ciphertext{}, fmt.Errorf("could not generate %d-byte nonce", chacha20poly1305.NonceSizeX)
	}
	if err != nil {
		return Ciphertext{}, err
	}
	ct := cipher.Seal([]byte{}, nonce, pt, []byte{})
	return Ciphertext{ct, nonce}, err
}

func Decrypt(sk Subkey, ct Ciphertext) ([]byte, error) {
	if len(sk) != SUBKEY_L {
		return []byte{}, fmt.Errorf("decryption key must be %d bytes", SUBKEY_L)
	}
	if len(ct.Nonce) != chacha20poly1305.NonceSizeX {
		return []byte{}, fmt.Errorf("nonce must be %d bytes", chacha20poly1305.NonceSizeX)
	}
	cipher, err := chacha20poly1305.NewX(sk)
	if err != nil {
		return []byte{}, err
	}
	pt, err := cipher.Open([]byte{}, ct.Nonce, ct.Data, []byte{})
	if err != nil {
		return []byte{}, err
	}
	return pt, err
}
