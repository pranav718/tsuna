package queue

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Item struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	AddedBy  string `json:"added_by"`
}

type Queue struct {
	items   []Item
	current int
	mu      sync.RWMutex
}

func New() *Queue {
	return &Queue{
		items:   make([]Item, 0),
		current: -1,
	}
}

func (q *Queue) Add(filename, addedBy string) Item {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := Item{
		ID:       generateID(),
		Filename: filename,
		AddedBy:  addedBy,
	}
	q.items = append(q.items, item)

	if q.current == -1 {
		q.current = 0
	}

	return item
}

func (q *Queue) Remove(itemID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if item.ID == itemID {
			q.items = append(q.items[:i], q.items[i+1:]...)
			if q.current >= len(q.items) {
				q.current = len(q.items) - 1
			}
			return true
		}
	}
	return false
}

func (q *Queue) Next() (Item, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	next := q.current + 1
	if next >= len(q.items) {
		return Item{}, false
	}
	q.current = next
	return q.items[q.current], true
}

func (q *Queue) Current() (Item, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.current < 0 || q.current >= len(q.items) {
		return Item{}, false
	}
	return q.items[q.current], true
}

func (q *Queue) Items() []Item {
	q.mu.RLock()
	defer q.mu.RUnlock()

	out := make([]Item, len(q.items))
	copy(out, q.items)
	return out
}

func (q *Queue) CurrentIndex() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.current
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
