package main

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Node struct {
	name     string
	previous *Node
	next     *Node
}

type List struct {
	count int32
	name  string
	head  *Node
	tail  *Node
}

func NewCreateList(name string) *List {
	return &List{
		name: name,
	}
}

func (l *List) AddName(name string) {
	newNode := &Node{
		name: name,
	}

	tail := l.tail
	if tail == nil {
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head)), nil, unsafe.Pointer(newNode)) {
			l.tail = newNode
			atomic.AddInt32(&l.count, 1)
			return
		}
	} else {
		next := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)))
		if next == nil {
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)), nil, unsafe.Pointer(newNode)) {
				atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&newNode.previous)), unsafe.Pointer(tail))
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&l.tail)), unsafe.Pointer(tail), unsafe.Pointer(newNode))
				atomic.AddInt32(&l.count, 1)
				return
			}
		} else {
			atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&l.tail)), unsafe.Pointer(tail), unsafe.Pointer(next))
		}
	}

}

func (l *List) ShowList() error {
	currentNode := l.head
	if currentNode == nil {
		return errors.New("WARNING: List is empty")
	}

	log.Printf("Count of elements - %d in %s", atomic.LoadInt32(&l.count), l.name)
	log.Printf("HEAD - %+v, address-[%p]", currentNode, currentNode)

	for currentNode.next != nil {
		currentNode = currentNode.next
		if currentNode != l.tail {
			log.Printf("%+v, address-[%p]", currentNode, currentNode)
		} else {
			log.Printf("TAIL - %+v, address-[%p]", currentNode, currentNode)
		}
	}

	return nil
}

// В данном пункте сделал, чтобы элементы шли не последовательно, а создавали условный плейлист со случайной последовательностью
func main() {
	l := NewCreateList("Random playlist")

	playlist := []string{"1", "2", "3", "4", "5", "6", "7"}

	var wg sync.WaitGroup

	wg.Add(len(playlist))
	for _, prince := range playlist {
		go func(p string) {
			defer wg.Done()
			l.AddName(p)
		}(prince)
	}

	wg.Wait()

	l.ShowList()

}
