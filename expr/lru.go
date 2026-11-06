package expr

import (
	"sync"
)

// DefaultLRUCapacity defines the maximum number of compiled expressions stored in cache.
const DefaultLRUCapacity = 4096

type lruNode struct {
	key   string
	val   Expr
	prev  *lruNode
	next  *lruNode
}

// lruCache is a thread-safe, bounded LRU cache for compiled expressions.
// It avoids unbounded memory growth when dynamic expressions are evaluated.
type lruCache struct {
	capacity int
	items    map[string]*lruNode
	head     *lruNode
	tail     *lruNode
	mu       sync.Mutex
}

func newLRUCache(capacity int) *lruCache {
	if capacity <= 0 {
		capacity = DefaultLRUCapacity
	}
	head := &lruNode{}
	tail := &lruNode{}
	head.next = tail
	tail.prev = head

	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*lruNode, capacity),
		head:     head,
		tail:     tail,
	}
}

func (c *lruCache) get(key string) (Expr, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, found := c.items[key]
	if !found {
		return nil, false
	}
	c.moveToFront(node)
	return node.val, true
}

func (c *lruCache) put(key string, val Expr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, found := c.items[key]; found {
		node.val = val
		c.moveToFront(node)
		return
	}

	if len(c.items) >= c.capacity {
		c.removeOldest()
	}

	newNode := &lruNode{
		key: key,
		val: val,
	}
	c.items[key] = newNode
	c.addToFront(newNode)
}

func (c *lruCache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func (c *lruCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*lruNode, c.capacity)
	c.head.next = c.tail
	c.tail.prev = c.head
}

func (c *lruCache) addToFront(node *lruNode) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

func (c *lruCache) removeNode(node *lruNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (c *lruCache) moveToFront(node *lruNode) {
	c.removeNode(node)
	c.addToFront(node)
}

func (c *lruCache) removeOldest() {
	oldest := c.tail.prev
	if oldest != c.head {
		c.removeNode(oldest)
		delete(c.items, oldest.key)
	}
}
