package pool

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	items []T
	newFn func() T
}

func New[T Resettable](newFn func() T) *Pool[T] {
	return &Pool[T]{
		items: make([]T, 0),
		newFn: newFn,
	}
}

func (p *Pool[T]) Get() T {
	n := len(p.items)
	if n == 0 {
		return p.newFn()
	}

	item := p.items[n-1]
	p.items = p.items[:n-1]
	return item
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.items = append(p.items, item)
}
