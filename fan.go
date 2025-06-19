package progressbar

import "sync"

func emptyIt(ch <-chan struct{}) (empty bool) {
	for {
		select {
		case <-ch:
			empty = false
		default:
			empty = true
			return
		}
	}
}

func fanIn[T any](bufSize int, inputs ...<-chan T) (out chan<- T) {
	out = make(chan<- T, bufSize)
	var wg sync.WaitGroup
	wg.Add(len(inputs))
	for _, in := range inputs {
		go func(in <-chan T) {
			defer wg.Done()
			for v := range in {
				out <- v
			}
		}(in)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return
}

func fanOut[T any](in <-chan T, n int) (outputs []chan<- T) {
	outputs = make([]chan<- T, n)
	for i := range n {
		outputs[i] = make(chan<- T)
	}
	go func() {
		for v := range in {
			for _, o := range outputs {
				o <- v
			}
		}
	}()
	return
}
