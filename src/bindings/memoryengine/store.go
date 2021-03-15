package memoryengine

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/genjidb/genji/engine"
)

// item implements an engine.Item.
// it is also used as a btree.Item.
type item struct {
	k, v []byte
	// set to true if the item has been deleted
	// during the current transaction
	// but before rollback or commit.
	deleted bool
}

func (i *item) Key() []byte {
	return i.k
}

func (i *item) ValueCopy(buf []byte) ([]byte, error) {
	if len(buf) < len(i.v) {
		buf = make([]byte, len(i.v))
	}
	n := copy(buf, i.v)
	return buf[:n], nil
}

// storeTx implements an engine.Store.
type storeTx struct {
	sl   *sortedList
	tx   *transaction
	name string
	ctx  context.Context
}

func (s *storeTx) Put(k, v []byte) error {
	select {
	case <-s.tx.ctx.Done():
		return s.tx.ctx.Err()
	default:
	}

	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	if len(k) == 0 {
		return errors.New("empty keys are forbidden")
	}

	if len(v) == 0 {
		return errors.New("empty values are forbidden")
	}

	it := &item{k: k}

	// if there is an existing value, fetch it
	// and overwrite it directly using the pointer.
	if i := s.sl.Get(it); i != nil {
		oldv, oldDeleted := i.v, i.deleted
		i.v = v
		i.deleted = false

		// on rollback replace the new value by the old value
		s.tx.onRollback = append(s.tx.onRollback, func() {
			i.v = oldv
			i.deleted = oldDeleted
		})

		return nil
	}

	it.v = v
	s.sl.Put(it)

	// on rollback delete the new item
	s.tx.onRollback = append(s.tx.onRollback, func() {
		s.sl.Delete(it)
	})

	return nil
}

func (s *storeTx) Get(k []byte) ([]byte, error) {
	select {
	case <-s.tx.ctx.Done():
		return nil, s.tx.ctx.Err()
	default:
	}

	it := s.sl.Get(&item{k: k})

	if it == nil {
		return nil, engine.ErrKeyNotFound
	}

	// don't return items that have been deleted during
	// this transaction.
	if it.deleted {
		return nil, engine.ErrKeyNotFound
	}

	return it.v, nil
}

// Delete marks k for deletion. The item will be actually
// deleted during the commit phase of the current transaction.
// The deletion is delayed to avoid a rebalancing of the tree
// every time we remove an item from it,
// which causes iterators to behave incorrectly when looping
// and deleting at the same time.
func (s *storeTx) Delete(k []byte) error {
	select {
	case <-s.tx.ctx.Done():
		return s.tx.ctx.Err()
	default:
	}

	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	it := s.sl.Get(&item{k: k})
	if it == nil {
		return engine.ErrKeyNotFound
	}

	// items that have been deleted during
	// this transaction must be ignored.
	if it.deleted {
		return engine.ErrKeyNotFound
	}

	// set the deleted flag to true.
	// this makes the item invisible during this
	// transaction without actually deleting it
	// from the tree.
	// once the transaction is commited, actually
	// remove it from the tree.
	it.deleted = true

	// on rollback set the deleted flag to false.
	s.tx.onRollback = append(s.tx.onRollback, func() {
		it.deleted = false
	})

	// on commit, remove the item from the tree.
	s.tx.onCommit = append(s.tx.onCommit, func() {
		if it.deleted {
			s.sl.Delete(it)
		}
	})
	return nil
}

// Truncate replaces the current tree by a new
// one. The current tree will be garbage collected
// once the transaction is commited.
func (s *storeTx) Truncate() error {
	select {
	case <-s.tx.ctx.Done():
		return s.tx.ctx.Err()
	default:
	}

	if !s.tx.writable {
		return engine.ErrTransactionReadOnly
	}

	old := s.sl
	s.sl = newSortedList()

	// on rollback replace the new tree by the old one.
	s.tx.onRollback = append(s.tx.onRollback, func() {
		s.sl = old
	})

	return nil
}

// NextSequence returns a monotonically increasing integer.
func (s *storeTx) NextSequence() (uint64, error) {
	select {
	case <-s.tx.ctx.Done():
		return 0, s.tx.ctx.Err()
	default:
	}

	if !s.tx.writable {
		return 0, engine.ErrTransactionReadOnly
	}

	s.tx.ng.sequences[s.name]++

	return s.tx.ng.sequences[s.name], nil
}

// Iterator creates an iterator with the given options.
func (s *storeTx) Iterator(opts engine.IteratorOptions) engine.Iterator {
	return &iterator{
		tx:      s.tx,
		sl:      s.sl,
		reverse: opts.Reverse,
		ctx:     s.ctx,
	}
}

// iterator uses a goroutine to read from the tree on demand.
type iterator struct {
	tx      *transaction
	reverse bool
	sl      *sortedList
	curIdx  int // current item
	ctx     context.Context
	err     error
}

func (it *iterator) Seek(pivot []byte) {
	select {
	case <-it.ctx.Done():
		it.err = it.ctx.Err()
		return
	default:
	}

	it.curIdx = -1
	if !it.reverse {
		if pivot != nil {
			it.curIdx = sort.SearchStrings(it.sl.sortedKeys, string(pivot))
		}
		it.Next()
		return
	}

	if pivot == nil {
		it.curIdx = len(it.sl.sortedKeys)
	} else {
		// slower O(n)
		spivot := string(pivot)
		for i := len(it.sl.sortedKeys) - 1; i >= 0; i-- {
			it.curIdx = i
			if strings.Compare(it.sl.sortedKeys[i], spivot) <= 0 {
				break
			}
		}
	}

	it.Next()
}

func (it *iterator) Valid() bool {
	if it.err != nil {
		return false
	}

	if it.reverse {
		return it.curIdx >= 0
	}

	return it.curIdx < len(it.sl.sortedKeys)
}

// Read the next item from the goroutine
func (it *iterator) Next() {
	for {
		if it.reverse {
			it.curIdx--
		} else {
			it.curIdx++
		}
		// fmt.Printf("Current index %d\n", it.curIdx)

		if !it.Valid() {
			return
		}
		// fmt.Printf("%#v\n", it.Item())

		if !it.Item().(*item).deleted {
			return
		}

		// fmt.Println("Skipped deleted")
	}
}

func (it *iterator) Err() error {
	return it.err
}

func (it *iterator) Item() engine.Item {
	return it.sl.items[it.sl.sortedKeys[it.curIdx]]
}

// Close the inner goroutine.
func (it *iterator) Close() error {
	return nil
}

type sortedList struct {
	items      map[string]*item
	sortedKeys []string
}

func newSortedList() *sortedList {
	return &sortedList{
		items: make(map[string]*item),
	}
}

func (s *sortedList) Put(it *item) {
	_, ok := s.items[string(it.k)]
	if !ok {
		key := string(it.k)
		s.sortedKeys = append(s.sortedKeys, key)
		sort.Strings(s.sortedKeys)
	}

	s.items[string(it.k)] = it
}

func (s *sortedList) Get(it *item) *item {
	return s.items[string(it.k)]
}

func (s *sortedList) Delete(it *item) *item {
	i, ok := s.items[string(it.k)]
	if !ok {
		return nil
	}

	delete(s.items, string(it.k))
	idx := sort.SearchStrings(s.sortedKeys, string(it.k))
	s.sortedKeys = append(s.sortedKeys[:idx], s.sortedKeys[idx+1:]...)
	return i
}
