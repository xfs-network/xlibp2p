package badger

import (
	"github.com/dgraph-io/badger/v3"
	"io/ioutil"
	"log"
)

type Storage struct {
	db *badger.DB
}

type loggingLevel int

const (
	DEBUG loggingLevel = iota
	INFO
	WARNING
	ERROR
)

type defaultLog struct {
	*log.Logger
	level loggingLevel
}

func (l *defaultLog) Errorf(f string, v ...interface{}) {
	if l.level <= ERROR {
		l.Printf("ERROR: "+f, v...)
	}
}

func (l *defaultLog) Warningf(f string, v ...interface{}) {
	if l.level <= WARNING {
		l.Printf("WARNING: "+f, v...)
	}
}

func (l *defaultLog) Infof(f string, v ...interface{}) {
	if l.level <= INFO {
		l.Printf("INFO: "+f, v...)
	}
}

func (l *defaultLog) Debugf(f string, v ...interface{}) {
	if l.level <= DEBUG {
		l.Printf("DEBUG: "+f, v...)
	}
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

func defaultLogger(level loggingLevel) *defaultLog {
	return &defaultLog{
		Logger: log.New(ioutil.Discard, "badger ", log.LstdFlags),
		level:  level,
	}
}

func New(pathname string) *Storage {
	storage := &Storage{}
	opts := badger.DefaultOptions(pathname)
	opts.Logger = defaultLogger(ERROR)
	var err error = nil
	storage.db, err = badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return storage
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
