package custom_channel

import "sync"

type Channel struct {
	// 互斥体
	mutex sync.Mutex
	// 条件变量
	cond *sync.Cond
	// 挂起队列
	queue *Queue
	// 挂起队列容量
	n int
}

func NewChannel(n int) *Channel {
	if n < 1 {
		panic("todo: support unbuffered channel")
	}
	c := new(Channel)
	c.cond = sync.NewCond(&c.mutex)
	c.queue = NewQueue()
	c.n = n
	return c
}

/**
 * Push
 * @param v interface{}为任务
 */
func (c *Channel) Push(v interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for c.queue.Len() == c.n {
		// 等待队列满，则立即等待
		c.cond.Wait()
	}

	// 等待队列为空，可能有人之前等待数据，通知它们继续
	if c.queue.Len() == 0 {
		c.cond.Broadcast()
	}
	c.queue.Push(v)
}

func (c *Channel) Pop() (v interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for c.queue.Len() == 0 {
		// 等待队列为空，则该操作阻塞
		c.cond.Wait()
	}

	if c.queue.Len() == c.n {
		// 等待队列空，可能之前等着写数据，通知他们
		c.cond.Broadcast()
	}

	return c.queue.Pop()
}

func (c *Channel) TryPop() (v interface{}, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.queue.Len() == 0 {
		return
	}

	if c.queue.Len() == c.n {
		c.cond.Broadcast()
	}
	return c.queue.Pop(), true
}

func (c *Channel) TryPush(v interface{}) (ok bool) {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.queue.Len() == c.n {
		return
	}

	// 等待队列为空，可能有人之前等待数据，通知它们继续
	if c.queue.Len() == 0 {
		c.cond.Broadcast()
	}
	c.queue.Push(v)
	return true
}
