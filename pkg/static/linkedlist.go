package static

import (
	"errors"
	"fmt"
	"reflect"
)

//链表中的数据类型
type ElementType int

//节点结构体
type INode struct {
	X    ElementType //节点
	next *INode      //下一节点指针
}

//链表结构体
type LinkedList struct {
	Head *INode //有一个头节点
}

func NewINode(x ElementType, next *INode) *INode {
	return &INode{x, next}
}
func NewLinkedList() *LinkedList {
	//头节点的数据域存储链表数据元素的大小（不包括头节点）
	head := &INode{0, nil}
	return &LinkedList{head}
}

/**
 * 判断链表是否为空
 */
func (list *LinkedList) IsEmpty() bool {
	return list.Head.next == nil
}

/**
 * 链表的长度
 */
func (list *LinkedList) Length() int {
	return int(reflect.ValueOf(list.Head.X).Int())
}

/**
  向链表后端Append元素
*/
func (list *LinkedList) Append(node *INode) {
	current := list.Head
	for {
		if current.next == nil {
			break
		}
		current = current.next
	}
	current.next = node
	list.sizeInc()
}

/**
  每次都在前端插入元素
*/
func (list *LinkedList) Prepend(node *INode) {
	current := list.Head
	node.next = current.next
	current.next = node
	list.sizeInc()
}

/**
  根据指定的数据查找*INode
*/
func (list *LinkedList) Find(x ElementType) (*INode, bool) {
	empty := list.IsEmpty()
	if empty {
		fmt.Println("This is an empty list")
		return nil, false
	}
	current := list.Head
	for current.next != nil {
		if current.X == x {
			return current, true
		}
		current = current.next
	}
	if current.X == x {
		return current, true
	}
	return nil, false
}

/**
  根据指定索引的数据查找*INode
*/
func (list *LinkedList) FindByIndex(index int) (*INode, bool) {
	empty := list.IsEmpty()
	if empty {
		fmt.Println("This is an empty list")
		return nil, false
	}
	if index >= list.Length() {
		fmt.Println("the index is too large")
		return nil, false
	}
	var i = 0
	current := list.Head.next
	pre := list.Head
	for current.next != nil {
		if i == index {
			pre.next = current.next
			list.sizeDec()
			return current, true
		}
		i++
		pre = current
		current = current.next
	}
	return nil, false
}

/**
  删除给定元素的节点
*/
func (list *LinkedList) Remove(x ElementType) error {
	empty := list.IsEmpty()
	if empty {
		return errors.New("This is an empty list")
	}
	current := list.Head
	for current.next != nil {
		if current.next.X == x {
			current.next = current.next.next
			list.sizeDec()
			return nil
		}
		current = current.next
	}
	return nil
}

/**
  内部函数，处理头节点记录的链表大小
*/
func (list *LinkedList) sizeInc() {
	v := int(reflect.ValueOf((*list.Head).X).Int())
	list.Head.X = ElementType(v + 1)
}

/**
  内部函数，处理头节点记录的链表大小
*/
func (list *LinkedList) sizeDec() {
	v := int(reflect.ValueOf((*list.Head).X).Int())
	list.Head.X = ElementType(v - 1)
}

/**
  打印链表信息
*/
func (list *LinkedList) PrintList() {
	empty := list.IsEmpty()
	if empty {
		fmt.Println("This is an empty list")
		return
	}
	current := list.Head.next
	fmt.Println("The elements is:")
	i := 0
	for ; ; i++ {
		if current.next == nil {
			break
		}
		fmt.Printf("INode%d ,value:%v -> ", i, current.X)
		current = current.next
	}
	fmt.Printf("Node%d value:%v", i+1, current.X)
	return

}
