package enet

type timers []*enet_command

// sort interface

func (self timers) Len() int           { return len(self) }
func (self timers) Less(i, j int) bool { return self[i].timeo < self[j].timeo }
func (self timers) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
	self[i].heap_idx = i
	self[j].heap_idx = j
}

// heap interface

func (self *timers) Push(x interface{}) {
	v := x.(*enet_command)
	v.heap_idx = len(*self)
	*self = append(*self, v)
}

func (self *timers) Pop() interface{} {
	l := len(*self)
	v := (*self)[l-1]
	*self = (*self)[:l-1]
	v.heap_idx = -1
	return v
}
