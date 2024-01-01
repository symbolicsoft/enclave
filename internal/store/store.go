// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package store

import (
	"errors"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"google.golang.org/protobuf/proto"

	enclaveProto "github.com/symbolicsoft/enclave/v2/internal/proto"
)

var Database = func() *leveldb.DB {
	db, err := leveldb.OpenFile("enclave.db", nil)
	if err != nil {
		log.Fatal("could not open database")
	}
	return db
}()

func HasNotebook(notebookId []byte) (bool, error) {
	return Database.Has(notebookId, &opt.ReadOptions{})
}

func PutNotebook(notebookId []byte, entry *enclaveProto.EncryptedNotebook, overwrite bool) error {
	var entryBytes []byte
	has, err := HasNotebook(notebookId)
	if err != nil {
		return err
	}
	if has && !overwrite {
		return errors.New("refusing to overwrite entry")
	}
	entryBytes, err = proto.Marshal(entry)
	if err != nil {
		return err
	}
	return Database.Put(notebookId, entryBytes, &opt.WriteOptions{Sync: true})
}

func GetNotebook(notebookId []byte) (*enclaveProto.EncryptedNotebook, error) {
	entryBytes, err := Database.Get(notebookId, &opt.ReadOptions{})
	if err != nil {
		return &enclaveProto.EncryptedNotebook{}, err
	}
	entry := &enclaveProto.EncryptedNotebook{}
	err = proto.Unmarshal(entryBytes, entry)
	if err != nil {
		return &enclaveProto.EncryptedNotebook{}, err
	}
	return entry, err
}

func DeleteNotebook(notebookId []byte) error {
	return Database.Delete(notebookId, &opt.WriteOptions{Sync: true})
}

func CloseDatabase() {
	err := Database.Close()
	if err != nil {
		log.Fatal("could not close database")
	}
}
