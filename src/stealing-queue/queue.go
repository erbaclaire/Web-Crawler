// Package queue creates a queue that can be stolen from
package queue

import (
	"sync/atomic"
	"unsafe"
)

// Queue interface represents a queue that can be stolen from with the following methods.
type Queue interface {
	PushBottom(bookUrl string)
    PopBottom() string
	PopTop() string
}

// queue is a slice of book urls that need to be processed.
// bottom represents the end of the slice
type queue struct {
	tasks []string
    bottom int
    top *top
}

// NewQueue initializes a new empty queue.
func NewQueue() Queue {
    q := new(queue)
    top := new(top)
    q.top = top
	return q
}

// top represents the index, stamp of an atomic stamped reference.
type top struct {
    index int
    stamp int
}

// PushBottom appends bookUrl's to the end of the queue's task slice.
func (q *queue) PushBottom(bookUrl string) {
    q.tasks = append(q.tasks, bookUrl)
    q.bottom++
}

// PopTop pops the book url at the queue's top index value.
// Uses the top strcuture as the unsafe Pointer in the CAS for ABA problem.
func (q *queue) PopTop() string {
    oldTop := q.top
    newTop := &top{q.top.index+1, q.top.stamp+1}
    if q.bottom <= oldTop.index {
        return "nil"
    } 
    t := q.tasks[oldTop.index]
    if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.top)), unsafe.Pointer(oldTop), unsafe.Pointer(newTop)) {
        return t
    }
    return "nil"
}

// PopBottom  pops the book url at the queue's last index value.
func (q *queue) PopBottom() string {
    // If nothing more to grab then return nil string.
    if q.bottom == 0 {
        return "nil"
    }
    q.bottom--
    t := q.tasks[q.bottom]
    oldTop := q.top
    newTop := &top{0, q.top.stamp+1}
    // If the bottom's valye is less than the top index's value then it is ok to dequeue.
    if q.bottom > oldTop.index {
        return t
    }
    // Otherwise, if bottom and top are equal, try to steal the last task.
    if q.bottom == oldTop.index {
        q.bottom = 0
        if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.top)), unsafe.Pointer(oldTop), unsafe.Pointer(newTop)) {
            return t
        }
        q.top.index = newTop.index 
        q.top.stamp = newTop.stamp
    }
    return "nil"    
}

// NewQueueSlice genereate a slice of newly initialized stealing queues.
// Each entry in the slice  will represent the stealing queue of one of the goroutines.
func NewQueueSlice(queues int) []Queue {
    var qs []Queue
    for i := 0; i < queues; i++ {
        q := NewQueue()
        qs = append(qs, q)
    }
    return qs
}