package main

// NewBlackHole returns a chan which dumps everything it receives
func NewBlackHole() chan *stat {
	ch := make(chan *stat, 1)
	go func() {
		for {
			_, open := <-ch
			if !open {
				return
			}
		}
	}()

	return ch
}
