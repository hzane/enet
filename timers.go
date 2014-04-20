package enet

type timer interface {
	run(Host)
}

type timer_item struct {
	index  int   // heap index
	weight int64 // unixtime
	value  timer
}

type timers []timer_item

// sort interface

func (self timers) Len() int           { return len(self) }
func (self timers) Less(i, j int) bool { return self[i].weight < self[j].weight }
func (self timers) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
	self[i].index = i
	self[j].index = j
}

// heap interface

func (self *timers) Push(x interface{}) {
	v := x.(timer_item)
	v.index = len(*self)
	*self = append(*self, v)
}

func (self *timers) Pop() interface{} {
	l := len(*self)
	v := (*self)[l-1]
	*self = (*self)[:l-1]
	return v
}
