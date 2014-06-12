package commiter

import (
	"errors"
	"fmt"
	"message"
)

type MemoryCommiter struct {
	buf     []interface{}
	maxSize int
	maxId   uint64
	curId   uint64
}

func NewMemoryCommiter() *MemoryCommiter {
	mSize := 1024
	b := make([]interface{}, 0, mSize)
	return &MemoryCommiter{buf: b, maxSize: mSize}
}

func (c *MemoryCommiter) Commit(msg interface{}) error {
	if len(c.buf) >= c.maxSize {
		println("Commit error: len(c.buf) = ", len(c.buf), " , c.maxSize = ", c.maxSize)
		return errors.New("MemoryCommiter buf size limited")
	}
	//println("commit msg ", msg)
	//println("maxId: ", c.maxId)
	//println("curId: ", c.curId)

	var newMsg message.Msg
	newMsg.MsgId = c.maxId
	newMsg.Content = msg
	c.buf = append(c.buf, newMsg)
	c.maxId++
	fmt.Println("Commit Msg: ", newMsg)
	return nil
}

func (c *MemoryCommiter) GetMsg(id uint64) (msg interface{}, err error) {
	if id < c.curId || id >= c.maxId {
		err = errors.New("Id out of range")
		return
	}
	msg = c.buf[id-c.curId]
	return
}

func (c *MemoryCommiter) Consume() (msg interface{}, err error) {
	if len(c.buf) == 0 {
		err = errors.New("No msg to consume")
		return
	}
	msg = c.buf[0]
	c.curId++
	c.buf = c.buf[1:]
	return
}

func (c *MemoryCommiter) Total() uint64 {
	return uint64(len(c.buf))
}

func (c *MemoryCommiter) CurId() uint64 {
	return c.curId
}

func (c *MemoryCommiter) MaxId() uint64 {
	return c.maxId
}
