package sqlsplit

import (
	"fmt"
	"sync"
)

// Item 为栈中容纳的任意元素接口类型
type Item interface{}

// ItemStack 为支持并发读写安全的通用元素栈结构体
type ItemStack struct {
	// items 存储栈内各元素切片（后进先出 LIFO）
	items []Item
	// lock 为读写互斥锁，保障并发安全
	lock sync.RWMutex
}

// NewStack 创建并返回一个空的并发安全栈实例
func NewStack() *ItemStack {
	s := &ItemStack{}
	s.items = []Item{}
	return s
}

// Print 打印输出当前栈内的所有元素（用于调试诊断）
func (s *ItemStack) Print() {
	fmt.Println(s.items)
}

// Push 将新元素 t 压入栈顶
func (s *ItemStack) Push(t Item) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.items = append(s.items, t)
}

// Pop 弹出并返回栈顶元素；若栈为空则返回 nil
func (s *ItemStack) Pop() Item {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 重要判断：若栈为空，返回 nil 防止越界切片访问
	if len(s.items) == 0 {
		return nil
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[0 : len(s.items)-1]
	return item
}

// Look 查看并返回当前栈顶元素但不弹出；若栈为空则返回 nil
func (s *ItemStack) Look() Item {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 重要判断：若栈为空，返回 nil 防止切片越界
	if len(s.items) == 0 {
		return nil
	}
	return s.items[len(s.items)-1]
}

// Len 返回当前栈中的元素数量
func (s *ItemStack) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items)
}
