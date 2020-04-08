package simpleengine

import (
	"errors"
	"sort"
	"strings"

	"github.com/asdine/genji/engine"
)

type item struct {
	k, v    []byte
	deleted bool
}

func (i *item) Key() []byte {
	return i.k
}

func (i *item) ValueCopy(b []byte) ([]byte, error) {
	return i.v, nil
}

type storeTx struct {
	*storeData

	tx *transaction
}

func (s *storeTx) Put(k, v []byte) error {
	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	if len(k) == 0 {
		return errors.New("empty keys are forbidden")
	}

	it, ok := s.values[string(k)]
	if ok {
		oldit := *it
		it.v = v
		it.deleted = false

		s.tx.onRollback = append(s.tx.onRollback, func() {
			s.values[string(k)] = &oldit
		})

		return nil
	}

	it = &item{
		k: k,
		v: v,
	}
	s.values[string(k)] = it
	s.keys = insertSorted(s.keys, string(k))

	s.tx.onRollback = append(s.tx.onRollback, func() {
		delete(s.values, string(k))
		s.keys = deleteSorted(s.keys, string(k))
	})

	return nil
}

func (s *storeTx) Get(k []byte) ([]byte, error) {
	it, ok := s.values[string(k)]
	if !ok || it.deleted {
		return nil, engine.ErrKeyNotFound
	}

	return it.v, nil
}

func (s *storeTx) Delete(k []byte) error {
	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	it, ok := s.values[string(k)]
	if !ok || it.deleted {
		return engine.ErrKeyNotFound
	}

	it.deleted = true

	s.tx.onRollback = append(s.tx.onRollback, func() {
		it.deleted = false
	})

	s.tx.onCommit = append(s.tx.onCommit, func() {
		delete(s.values, string(k))
		s.keys = deleteSorted(s.keys, string(k))
	})
	return nil
}

func (s *storeTx) Truncate() error {
	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	oldValues := s.values
	oldKeys := s.keys[:]

	s.values = make(map[string]*item)
	s.keys = nil

	s.tx.onRollback = append(s.tx.onRollback, func() {
		s.values = oldValues
		s.keys = oldKeys
	})

	return nil
}

func (s *storeTx) NewIterator(cfg engine.IteratorConfig) engine.Iterator {
	return &itemIterator{
		reverse: cfg.Reverse,
		keys:    s.keys,
		values:  s.values,
	}
}

type itemIterator struct {
	reverse bool
	keys    []string
	values  map[string]*item
	index   int
}

func (it *itemIterator) Seek(k []byte) {
	kk := string(k)
	if it.reverse {
		if len(k) == 0 {
			it.index = len(it.keys) - 1
			return
		}
		it.index = sort.Search(len(it.keys), func(i int) bool { return strings.Compare(it.keys[i], kk) >= 0 })
		if it.keys[it.index] != kk && it.index > 0 {
			it.index--
		}
		return
	}
	if len(k) == 0 {
		it.index = 0
		return
	}
	it.index = sort.Search(len(it.keys), func(i int) bool { return strings.Compare(it.keys[i], kk) >= 0 })
}

func (it *itemIterator) Next() {
	if it.reverse {
		for it.index-1 >= -1 {
			it.index--
			if it.index >= 0 && !it.values[string(it.keys[it.index])].deleted {
				return
			}
		}
		return
	}

	it.index++
	for it.index < len(it.keys) && it.values[string(it.keys[it.index])].deleted {
		it.index++
	}
}

func (it *itemIterator) Valid() bool {
	return it.index >= 0 && it.index < len(it.keys)
}

func (it *itemIterator) Item() engine.Item {
	return it.values[string(it.keys[it.index])]
}

func (it *itemIterator) Close() error {
	return nil
}

func insertSorted(keys []string, el string) []string {
	index := sort.Search(len(keys), func(i int) bool { return strings.Compare(keys[i], el) > 0 })
	keys = append(keys, "")
	copy(keys[index+1:], keys[index:])
	keys[index] = el
	return keys
}

func deleteSorted(keys []string, el string) []string {
	if len(keys) == 0 {
		return keys
	}
	index := sort.Search(len(keys), func(i int) bool { return strings.Compare(keys[i], el) >= 0 })
	if index >= len(keys) || keys[index] != el {
		return keys
	}

	copy(keys[index:], keys[index+1:])
	return keys[:len(keys)-1]
}
