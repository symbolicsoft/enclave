// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/symbolicsoft/enclave/v2/internal/ciphers"
	"github.com/symbolicsoft/enclave/v2/internal/notebook"
	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
	"github.com/symbolicsoft/enclave/v2/internal/store"
	"golang.org/x/crypto/chacha20poly1305"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type EnclaveServer struct {
	CertFilePath  string
	KeyFilePath   string
	ListenAddress string
	ListenPort    int
	enclaveProto.UnimplementedEnclaveServiceServer
}

func (es *EnclaveServer) Start() {
	creds, err := credentials.NewServerTLSFromFile(es.CertFilePath, es.KeyFilePath)
	if err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", es.ListenAddress, es.ListenPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	enclaveProto.RegisterEnclaveServiceServer(grpcServer, &EnclaveServer{})
	grpcServer.Serve(lis)
}

func (es *EnclaveServer) PingPong(ctx context.Context, pingRequest *enclaveProto.Ping) (*enclaveProto.Ping, error) {
	if len(pingRequest.Msg) != 8 {
		return &enclaveProto.Ping{}, errors.New("invalid ping message length")
	}
	return pingRequest, nil
}

func (es *EnclaveServer) PutNotebook(ctx context.Context, enb *enclaveProto.EncryptedNotebook) (*enclaveProto.PutNotebookResponse, error) {
	if len(enb.NotebookId) != ciphers.SUBKEY_L {
		return &enclaveProto.PutNotebookResponse{ResponseCode: 400}, errors.New("invalid notebook id")
	}
	if len(enb.Nonce) != chacha20poly1305.NonceSizeX {
		return &enclaveProto.PutNotebookResponse{ResponseCode: 400}, errors.New("invalid nonce")
	}
	if len(enb.Data) > notebook.NOTEBOOK_BYTES_MAX {
		return &enclaveProto.PutNotebookResponse{ResponseCode: 400}, errors.New("invalid notebook size")
	}
	if len(enb.DecoyFor) != 0 {
		if len(enb.DecoyFor) != ciphers.SUBKEY_L {
			return &enclaveProto.PutNotebookResponse{ResponseCode: 400}, errors.New("invalid notebook id")
		}
	} else {
		enb_, err := store.GetNotebook(enb.NotebookId)
		if err == nil {
			enb.DecoyFor = enb_.DecoyFor
			enb.DecoyFuse = enb_.DecoyFuse
		}
	}
	err := store.PutNotebook(enb.NotebookId, enb, true)
	if err != nil {
		return &enclaveProto.PutNotebookResponse{ResponseCode: 500}, errors.New("notebook storage failed")
	}
	return &enclaveProto.PutNotebookResponse{ResponseCode: 200}, nil
}

func (es *EnclaveServer) GetNotebook(ctx context.Context, notebookId *enclaveProto.NotebookId) (*enclaveProto.GetNotebookResponse, error) {
	if len(notebookId.Id) != ciphers.SUBKEY_L {
		return &enclaveProto.GetNotebookResponse{ResponseCode: 400}, errors.New("invalid notebook id")
	}
	has, _ := store.HasNotebook(notebookId.Id)
	if !has {
		time.Sleep(time.Second * 5)
	}
	enb, err := store.GetNotebook(notebookId.Id)
	if err != nil {
		return &enclaveProto.GetNotebookResponse{ResponseCode: 500}, errors.New("notebook retrieval failed")
	}
	if len(enb.Nonce) != chacha20poly1305.NonceSizeX {
		return &enclaveProto.GetNotebookResponse{ResponseCode: 400}, errors.New("invalid nonce")
	}
	if len(enb.Data) > notebook.NOTEBOOK_BYTES_MAX {
		store.DeleteNotebook(notebookId.Id)
		return &enclaveProto.GetNotebookResponse{ResponseCode: 400}, errors.New("invalid notebook size")
	}
	if len(enb.DecoyFor) != 0 {
		if len(enb.DecoyFor) != ciphers.SUBKEY_L {
			return &enclaveProto.GetNotebookResponse{ResponseCode: 400}, errors.New("invalid notebook id")
		}
		if enb.DecoyFuse {
			store.DeleteNotebook(enb.DecoyFor)
			enb.DecoyFor = []byte{}
			enb.DecoyFuse = false
		} else {
			enb.DecoyFuse = true
		}
		store.PutNotebook(notebookId.Id, enb, true)
	}
	return &enclaveProto.GetNotebookResponse{
		ResponseCode: 200,
		NotebookId:   enb.NotebookId,
		Data:         enb.Data,
		Nonce:        enb.Nonce,
	}, nil
}
