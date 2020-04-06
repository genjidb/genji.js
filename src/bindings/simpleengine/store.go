package simpleengine

import (
	"bytes"
	"context"
	"errors"

	"github.com/asdine/genji/engine"
	"github.com/google/btree"
)

type item struct {
	k, v    []byte
	deleted bool
}

func (i *item) Less(than btree.Item) bool {
	return bytes.Compare(i.k, than.(*item).k) < 0
}

func (i *item) Key() []byte {
	return i.k
}

func (i *item) ValueCopy(b []byte) ([]byte, error) {
	return i.v, nil
}

type storeTx struct {
	tr *btree.BTree
	tx *transaction
}

func (s *storeTx) Put(k, v []byte) error {
	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	if len(k) == 0 {
		return errors.New("empty keys are forbidden")
	}

	it := &item{k: k}
	if i := s.tr.Get(it); i != nil {
		cur := i.(*item)

		oldv, oldDeleted := cur.v, cur.deleted
		cur.v = v
		cur.deleted = false

		s.tx.onRollback = append(s.tx.onRollback, func() {
			cur.v = oldv
			cur.deleted = oldDeleted
		})

		return nil
	}

	it.v = v
	s.tr.ReplaceOrInsert(it)

	s.tx.onRollback = append(s.tx.onRollback, func() {
		s.tr.Delete(it)
	})

	return nil
}

func (s *storeTx) Get(k []byte) ([]byte, error) {
	it := s.tr.Get(&item{k: k})

	if it == nil {
		return nil, engine.ErrKeyNotFound
	}

	i := it.(*item)
	if i.deleted {
		return nil, engine.ErrKeyNotFound
	}

	return it.(*item).v, nil
}

func (s *storeTx) Delete(k []byte) error {
	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	it := s.tr.Get(&item{k: k})
	if it == nil {
		return engine.ErrKeyNotFound
	}

	i := it.(*item)
	if i.deleted {
		return engine.ErrKeyNotFound
	}

	i.deleted = true

	s.tx.onRollback = append(s.tx.onRollback, func() {
		i.deleted = false
	})

	s.tx.onCommit = append(s.tx.onCommit, func() {
		s.tr.Delete(i)
	})
	return nil
}

func (s *storeTx) Truncate() error {
	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	old := s.tr
	s.tr = btree.New(3)

	s.tx.onRollback = append(s.tx.onRollback, func() {
		s.tr = old
	})

	return nil
}

func (s *storeTx) NewIterator(cfg engine.IteratorConfig) engine.Iterator {
	return &itemIterator{
		reverse: cfg.Reverse,
		tr:      s.tr,
	}
}

type itemIterator struct {
	cancel  func()
	reverse bool
	tr      *btree.BTree
	ch      chan *item
	itm     *item
}

func (it *itemIterator) Seek(k []byte) {
	if it.cancel != nil {
		it.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	it.cancel = cancel

	it.ch = make(chan *item)

	iterator := btree.ItemIterator(func(i btree.Item) bool {
		itm := i.(*item)
		if itm.deleted {
			return true
		}

		select {
		case <-ctx.Done():
			return false
		case it.ch <- itm:
		}

		return true
	})

	go func() {
		defer close(it.ch)

		if it.reverse {
			if len(k) == 0 {
				it.tr.Descend(iterator)
			} else {
				it.tr.DescendLessOrEqual(&item{k: k}, iterator)
			}
		} else {
			if len(k) == 0 {
				it.tr.Ascend(iterator)
			} else {
				it.tr.AscendGreaterOrEqual(&item{k: k}, iterator)
			}
		}
	}()

	it.Next()
}

func (it *itemIterator) Next() {
	it.itm = <-it.ch
}

func (it *itemIterator) Valid() bool {
	if it.itm != nil {
		return it.itm.k != nil
	}

	return false
}

func (it *itemIterator) Item() engine.Item {
	return it.itm
}

func (it *itemIterator) Close() error {
	if it.cancel != nil {
		it.cancel()
	}

	return nil
}
