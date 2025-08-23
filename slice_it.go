package go_raptor

type SliceIterator[T any] struct {
	data    []T
	length  int
	index   int
	reverse bool
}

func NewSliceIterator[T any](data []T, reverse bool) *SliceIterator[T] {
	it := &SliceIterator[T]{data: data, length: len(data), reverse: reverse}
	if reverse {
		it.index = len(data) - 1
	} else {
		it.index = 0
	}
	return it
}

func (it *SliceIterator[T]) Length() int {
	return it.length
}

/**
 * Checks if the iterator can continue depending on the direction
 */
func (it *SliceIterator[T]) HasNext() bool {
	if it.reverse {
		return it.index >= 0
	}
	return it.index < len(it.data)
}

/**
 * gets the next item depending on the direction
 */
func (it *SliceIterator[T]) Next() T {
	if !it.HasNext() {
		panic("Next always has to be pre-guarded by HasNext")
	}

	val := it.data[it.index]

	if it.reverse {
		it.index--
	} else {
		it.index++
	}

	return val
}

/**
 * Get's the firs item based on the direction (could be the last one if reverse)
 */
func (it *SliceIterator[T]) First() T {
	if it.length == 0 {
		panic("can not get First element from empty slice")
	}

	if it.reverse {
		return it.data[it.length-1]
	}
	return it.data[0]
}

func (it *SliceIterator[T]) SliceIterator(from_inclusive int, num_items int) *SliceIterator[T] {
	if it.reverse {
		/* interpret from_inclusive index as the index as if it the list were in reverse */
		return NewSliceIterator(it.data[it.length-from_inclusive-num_items:it.length-from_inclusive], it.reverse)
	}
	return NewSliceIterator(it.data[from_inclusive:from_inclusive+num_items], it.reverse)
}

/**
 * Resets the iterator to 0
 */
func (it *SliceIterator[T]) Reset() {
	if it.reverse {
		it.index = it.length - 1
	} else {
		it.index = 0
	}
}
