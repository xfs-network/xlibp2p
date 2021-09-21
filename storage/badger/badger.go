// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package badger

import (
	"bytes"
	"encoding/binary"
	"github.com/dgraph-io/badger/v3"
	"os"
)

type Storage struct {
	db *badger.DB
	version uint32
}

type loggingLevel int

const (
	DEBUG loggingLevel = iota
	INFO
	WARNING
	ERROR
)

var versionKey = []byte("version")

type defaultLog struct {
}

func (l *defaultLog) Errorf(f string, v ...interface{}) {
}

func (l *defaultLog) Warningf(f string, v ...interface{}) {
}

func (l *defaultLog) Infof(f string, v ...interface{}) {
}

func (l *defaultLog) Debugf(f string, v ...interface{}) {
}

type StorageWriteBatch struct {
	batch *badger.WriteBatch
}

func (b *StorageWriteBatch) Put(key, value []byte) error {
	k := append([]byte{}, key...)
	v := append([]byte{}, value...)

	return b.batch.Set(k, v)
}

func (b *StorageWriteBatch) Clear() {
	panic("not supported")
}

func (b *StorageWriteBatch) Count() int {
	panic("not supported")
}

func (b *StorageWriteBatch) Destroy() {
	b.batch.Cancel()
}

func (b *StorageWriteBatch) Delete(key []byte) error {
	return b.batch.Delete(key)
}
func New(pathname string) *Storage {
	storage,err := NewByVersion(pathname, 0)
	if err != nil {
		panic(err)
	}
	return storage
}
func NewByVersion(pathname string, version uint32) (*Storage, error) {
	storage := &Storage{
		version: version,
	}
	opts := badger.DefaultOptions(pathname)
	opts.Logger = &defaultLog{}
	var err error = nil
	storage.db, err = badger.Open(opts)
	if err != nil {
		panic(err)
	}
	var currentVer [4]byte
	binary.LittleEndian.PutUint32(currentVer[:], version)
	gotVersion, _ := storage.GetData(versionKey)
	if gotVersion == nil {
		if err := storage.SetData(versionKey, currentVer[:]); err != nil {
			if err := storage.Close(); err != nil {
				panic(err)
			}
			return nil, err
		}
	} else if bytes.Compare(gotVersion, currentVer[:]) != 0 {
		if err := storage.Close(); err != nil {
			panic(err)
		}
		err := os.RemoveAll(pathname)
		if err != nil {
			return nil, err
		}
		return NewByVersion(pathname, version)
	}
	return storage, nil
}

func (storage *Storage) Set(key string, val []byte) error {
	return storage.SetData([]byte(key), val)
}

func (storage *Storage) SetData(key []byte, val []byte) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		return err
	})
}

func (storage *Storage) NewWriteBatch() *StorageWriteBatch {
	return &StorageWriteBatch{
		batch: storage.db.NewWriteBatch(),
	}
}
func (storage *Storage) CommitWriteBatch(batch *StorageWriteBatch) error {
	return batch.batch.Flush()
}

func (storage *Storage) Get(key string) ([]byte, error) {
	return storage.GetData([]byte(key))
}
func (storage *Storage) GetData(key []byte) (val []byte, err error) {
	err = storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(val)
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (storage *Storage) Del(key string) error {
	return storage.DelData([]byte(key))
}

func (storage *Storage) DelData(key []byte) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}


func (storage *Storage) Close() error {
	return storage.db.Close()
}

func (storage *Storage) Foreach(fn func(k string, v []byte) error) error {
	return storage.ForeachData(func(k []byte, v []byte) error {
		return fn(string(k), v)
	})
}

func (storage *Storage) ForeachData(fn func(k []byte, v []byte) error) error {
	return storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		//opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = fn(key, val)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (storage *Storage) For(fn func(k []byte,v []byte) ) {
	if err := storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		//opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			fn(key, val)
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (storage *Storage) ForIndex(fn func(n int, k []byte,v []byte) ) {
	if err := storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		//opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		i := 0
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			fn(i,key, val)
			i+=1
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
func (storage *Storage) ForIndexStar(start int,fn func(n int, k []byte,v []byte) ) {
	if err := storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		i := 0
		for it.Rewind(); it.Valid(); it.Next() {
			if i < start {
				continue
			}
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			fn(i,key, val)
			i+=1
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
func (storage *Storage) PrefixForeach(prefix string,fn func(k string,v []byte) error ) error {
	return storage.PrefixForeachData([]byte(prefix), func(k []byte, v []byte) error {
		return fn(string(k), v)
	})
}

func (storage *Storage) PrefixForeachData(prefix []byte, fn func(k []byte, v []byte) error) error {
	return storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = fn(key, val)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type Iterator interface {
	Next() bool
	Key() []byte
	Val() []byte
	Close()
}

type dbIterator struct {
	it *badger.Iterator
	txn *badger.Txn
	current *badger.Item
}

func (it *dbIterator) Next() bool {
	if !it.it.Valid() {
		return false
	}
	it.current = it.it.Item()
	it.it.Next()
	return true
}

func (it *dbIterator) Key() []byte {
	return it.current.Key()
}
func (it *dbIterator) Val() []byte {
	val,err := it.current.ValueCopy(nil)
	if err != nil {
		return nil
	}
	return val
}
func (it *dbIterator) Close() {
	it.it.Close()
	it.txn.Discard()
}

func (storage *Storage) NewIterator() Iterator {
	mTxn := storage.db.NewTransaction(true)
	opts := badger.DefaultIteratorOptions
	mIt := mTxn.NewIterator(opts)
	mIt.Rewind()
	return &dbIterator{
		it: mIt,
		txn: mTxn,
	}
}

func (storage *Storage) GetVersion() uint32 {
	return storage.version
}