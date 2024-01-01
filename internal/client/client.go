// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"

	"github.com/symbolicsoft/enclave/v2/internal/ciphers"
	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const SERVER_GRPC = "enclave.sh:7070"
const SERVER_CERT = `-----BEGIN CERTIFICATE-----
MIIBljCCATygAwIBAgIUCk9YeSAFCPRg0SQbi9leZHhVZOwwCgYIKoZIzj0EAwIw
FTETMBEGA1UEAwwKZW5jbGF2ZS5zaDAeFw0yMzEyMjgxOTQ4NDZaFw0yODEyMjYx
OTQ4NDZaMBUxEzARBgNVBAMMCmVuY2xhdmUuc2gwWTATBgcqhkjOPQIBBggqhkjO
PQMBBwNCAATQnIf1Wj1a8ajvO/nFeEbaE+aZUw+sgMDPZ4/82mYW62DLYuOg9y8t
KDiwCjDZm9hIqko00Vpjfy+11uUCUXpuo2owaDAdBgNVHQ4EFgQUSiIpa41KUdW2
o9FpjnyfJZ3Zz8MwHwYDVR0jBBgwFoAUSiIpa41KUdW2o9FpjnyfJZ3Zz8MwDwYD
VR0TAQH/BAUwAwEB/zAVBgNVHREEDjAMggplbmNsYXZlLnNoMAoGCCqGSM49BAMC
A0gAMEUCIQCdR/CV7cE+JhpCTrZyG/OaGM9MnaIIhiTFn0lSViqlgAIgWyCqYYq4
kQB4C/4Kwm4HmskeKBTx6z9ftCOr6qqpROM=
-----END CERTIFICATE-----`

func getClient() (*grpc.ClientConn, error) {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM([]byte(SERVER_CERT)) {
		return &grpc.ClientConn{}, errors.New("credentials: failed to append certificates")
	}
	return grpc.Dial(SERVER_GRPC, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: cp})))
}

func PingPong() error {
	conn, err := getClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	grpcClient := enclaveProto.NewEnclaveServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ping := &enclaveProto.Ping{
		Msg: make([]byte, 8),
	}
	n, err := rand.Read(ping.Msg)
	if n != 8 {
		return errors.New("could not generate ping bytes")
	}
	if err != nil {
		return err
	}
	pong, err := grpcClient.PingPong(ctx, ping)
	if err != nil {
		return err
	}
	if !bytes.Equal(ping.Msg, pong.Msg) {
		return errors.New("ping mismatch")
	}
	return nil
}

func PutNotebook(uskId ciphers.Subkey, decoyFor ciphers.Subkey, ct ciphers.Ciphertext) error {
	enb := &enclaveProto.EncryptedNotebook{
		NotebookId: uskId,
		DecoyFor:   decoyFor,
		DecoyFuse:  false,
		Data:       ct.Data,
		Nonce:      ct.Nonce,
	}
	conn, err := getClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	grpcClient := enclaveProto.NewEnclaveServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err = grpcClient.PutNotebook(ctx, enb)
	if err != nil {
		return err
	}
	return nil
}

func GetNotebook(uskId ciphers.Subkey) (*enclaveProto.EncryptedNotebook, error) {
	conn, err := getClient()
	if err != nil {
		return &enclaveProto.EncryptedNotebook{}, err
	}
	if err != nil {
		return &enclaveProto.EncryptedNotebook{}, err
	}
	defer conn.Close()
	grpcClient := enclaveProto.NewEnclaveServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	enb, err := grpcClient.GetNotebook(ctx, &enclaveProto.NotebookId{
		Id: uskId,
	})
	if err != nil {
		return &enclaveProto.EncryptedNotebook{}, err
	}
	return &enclaveProto.EncryptedNotebook{
		NotebookId: enb.NotebookId,
		DecoyFor:   []byte{},
		DecoyFuse:  false,
		Data:       enb.Data,
		Nonce:      enb.Nonce,
	}, nil
}
