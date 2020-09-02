// The linkedlist package is a lock-free linkedlist, to be used as a dictionary.
// Source: The Art of Multiprocessor Programming, pp. 234-239 - but no removes so much easier.
package linkedlist

import (
	"unsafe"
	"strconv"
	"strings"
	"sync/atomic"
	"fmt"
	"src/book"
)

// List is the interfaces for a lock-free linked list.
// Note there is no Remove method because I do not need this for my project
// and having it would make the program extremely complicated. 
type List interface {
	Add(key string, value *book.Book)
	Print()
}

// A node represents a key, value pair like a dictionary except the value is mutable. 
type node struct {
	key string
	value *book.Book
	next *node
}

// NewNode returns a new node structure with the given key and value pair.
func newNode(key string, value *book.Book, next *node) *node {
	return &node{key, value, next}
}

// A list represents a linked list.
type list struct {
	 start *node
}

// NewList creates a empty linked list.
// Dummy book data because it is not important.
func NewList() List {
	initList := newNode("0", nil, newNode("zzzzzzzzzzzzzzzzzz", nil, nil))
	return &list{start: initList}
}

// Add adds a new node to the linked list (lock free).
func (l *list) Add(key string, value *book.Book) {
	for {

		pred := l.start
		curr := l.start.next	

		// Find where node is supposed to be.
		for curr.key < key {
			pred = curr
			curr = curr.next
		}

		// If node is already in the list, then update it (atomically) if it is a "better" book.
		// No remove so no chance it will be removed before it can be updated.
		// Only have to worry about concurrent updates.
		newNode := newNode(key, value, curr)
		if curr.key == key {
			newNode.Update(curr)
			break
		}

		// Otherwise, when get to correct place in list, try to add the node.
		// If fail then try again. Fails could happen because another add happened concurrently.
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&pred.next)), unsafe.Pointer(curr), unsafe.Pointer(newNode)) {
			break
		}
	}
}

// Update updates a node's value in the linked list.
// Update if the book's price is less than the current listed book in the genre's price.
func (newNode *node) Update(currNode *node) {
	for {
		expectNodeValue := currNode.value
		currValPrice := currNode.value.Price
		newValPrice := newNode.value.Price
		if newValPrice < currValPrice {
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&currNode.value)), unsafe.Pointer(expectNodeValue), unsafe.Pointer(newNode.value)) {
				break
			}
		} else {
			return
		}
	}
	return
}

// Print allows visualization of the linked list.
func (l *list) Print() {
	node := l.start.next
	result := ""
	for node.key != "zzzzzzzzzzzzzzzzzz" {
	var inStockString string
		if node.value.InStock {
			inStockString = "Yes"
		} else {
			inStockString = "No"
		}
		result += strings.ToUpper(node.key) + "\nTitle: " + node.value.Title + "\nDescription:" + node.value.Description + "\nPrice: " + fmt.Sprintf("%.2f", node.value.Price) + " Pound Sterling\nIn Stock?: " + inStockString + "\nStars: " + strconv.Itoa(node.value.Stars) + "\n\n"
		node = node.next
	}
	fmt.Println(result)
}