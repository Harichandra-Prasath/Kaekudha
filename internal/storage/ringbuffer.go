package storage

type RingBuffer[T any] struct {
	buf      []T
	head     int
	tail     int
	size     int
	capacity int
}

func New[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buf:      make([]T, capacity),
		capacity: capacity,
	}
}

func (R *RingBuffer[T]) Push(v T) bool {
	if R.size == R.capacity {
		// drop the packets
		return false
	}

	R.buf[R.head] = v
	R.head = (R.head + 1) % R.capacity
	R.size += 1

	return true
}

func (R *RingBuffer[T]) Pop() (T, bool) {
	if R.size == 0 {
		var zero T
		return zero, false
	}

	v := R.buf[R.tail]
	var zero T
	R.buf[R.tail] = zero

	R.tail = (R.tail + 1) % R.capacity
	R.size -= 1

	return v, true
}

func (R *RingBuffer[T]) Len() int {
	return R.size
}
