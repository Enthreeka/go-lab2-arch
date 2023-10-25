package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

//	Реализуйте конкурентный двухсвязный список
//	a.Минимизировав количество блокировок
//	b.Не используя блокировку (Используя атомарные ссылки)

type Node struct {
	name     string
	previous *Node
	next     *Node
}

type List struct {
	count int
	name  string
	head  *Node
	tail  *Node

	mu sync.Mutex
}

// CreateList можно назвать конструктором для нашей структуры, через него мы сможем обращаться к методам
func CreateList(name string) *List {
	return &List{
		name: name,
	}
}

// AddName добавляет элементы в двусвязный список
func (l *List) AddName(name string) {
	// Используем мьютекс для избежания гонки данных при обращении нескольких потоков
	l.mu.Lock()
	defer l.mu.Unlock()

	// Добавляем в структуру данные, которые мы передали
	n := &Node{
		name: name,
	}
	// Проверяем, если ли вообще в двсвязном списке данные, если нет, то структура будеет HEAD
	if l.head == nil {
		l.head = n
		// В случае, если в списке уже есть данные, то мы получаем значение последнего элемента в списке
		// и делаем ему ссылку на новый элемент, которой является поступившая структура,
		// а уже у поступившей структуры даем ей ссылку на предедыщий элемент, который был TAIL
	} else {
		currentNode := l.tail
		currentNode.next = n
		n.previous = l.tail
	}
	//Делаем хвостом поступившую структуру
	l.tail = n
	l.count++
}

// Remove деает поиск ноды в списке и удаляет ее
func Remove(list *List, value string) error {
	if list == nil {
		return errors.New("list is empty")
	}

	currentNode := list.head
	for currentNode != nil {
		if currentNode.name == value {

			if currentNode.previous != nil {
				currentNode.previous.next = currentNode.next
			} else {
				list.head = currentNode.next
			}

			if currentNode.next != nil {
				currentNode.next.previous = currentNode.previous
			} else {
				list.tail = currentNode.previous
			}

			list.count--

			return nil
		}
		currentNode = currentNode.next
	}

	return errors.New("value not found in the list")
}

// LookUp делает поиск ноды в листе
func LookUp(list *List, value string) (*Node, error) {
	if list == nil {
		return nil, errors.New("list is empty")
	}

	currentNode := list.head
	for currentNode != nil {
		if currentNode.name == value {
			return currentNode, nil
		}
		currentNode = currentNode.next
	}

	return nil, errors.New("value not found in the list")
}

// MergeList объединяет несколько списков в один
func MergeList(lists ...*List) *List {
	if lists == nil {
		return nil
	}

	mergedList := CreateList("Russian princes merged")

	for i, list := range lists {
		if i == len(lists) {
			break
		}

		if i == 0 {
			mergedList.head = list.head
			mergedList.tail = list.tail
		} else {
			mergedList.tail.next = list.head
			list.head.previous = mergedList.tail
			mergedList.tail = list.tail
		}

		mergedList.count += list.count
	}

	return mergedList
}

// ShowList выводит все элементы списка
func (l *List) ShowList() error {

	currentNode := l.head
	if currentNode == nil {
		return errors.New("WARNING: List is empty")
	}

	log.Printf("Count of elements - %d in %s list", l.count, l.name)
	log.Printf("HEAD - %+v,address-[%p]\n", currentNode, currentNode)
	for currentNode.next != nil {
		currentNode = currentNode.next
		if currentNode != l.tail {
			log.Printf("%+v,address-[%p]\n", currentNode, currentNode)
		} else {
			log.Printf("TAIL - %+v,address-[%p]\n", currentNode, currentNode)
		}
	}

	return nil
}

// Возможно некорректно понял задачу, но в данном решении имеется 2 массива,в которых горутины
// конкурентно создают 2 разных списка. Потом через функцию MergeList объединяю эти списки.

// Исользуется семфор в виде канала sem, для записи данных в лист в порядке их очереди в массиве.
// Также ниже есть 2 решение v1.0, где всего 2 горутины, внутри которых уже работают циклы и создают листы.
func main() {
	l1 := CreateList("Russian princes part 1")
	l2 := CreateList("Russian princes part 2")

	princesFirstCentury := []string{"Олег Вещий", "Игорь Рюрикович", "Ольга", "Святослав Игоревич", "Ярополк Святославич"}
	princessSecondCentury := []string{"Владимир Святославич", "Святополк Владимирович", "Ярослав Владимирович Мудрый"}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 1)

	for _, prince := range princesFirstCentury {
		wg.Add(1)
		sem <- struct{}{}
		go func(prince string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			l1.AddName(prince)
		}(prince)
	}

	for _, prince := range princessSecondCentury {
		wg.Add(1)
		sem <- struct{}{}
		go func(prince string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			l2.AddName(prince)
		}(prince)
	}

	wg.Wait()

	l := MergeList(l1, l2)
	if l == nil {
		fmt.Println("nil-pointer in merged slice")
	}

	l.ShowList()

	//node, err := LookUp(l, "Ольга")
	//if err != nil {
	//	fmt.Println(err)
	//}

	//fmt.Printf("%#v\n", node)

	Remove(l, "Ольга")

	l.ShowList()
}

// a) v1.0 - Способ, где всего 2 потока и внутри каждого цикл со слайсами
//wg.Add(1)
//go func() {
//	defer wg.Done()
//
//	for _, prince := range princesFirstCentury {
//		l1.AddName(prince)
//	}
//
//}()
//
//wg.Add(1)
//go func() {
//	defer wg.Done()
//
//	for _, prince := range princessSecondCentury {
//		l2.AddName(prince)
//	}
//
//}()
