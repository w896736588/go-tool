package gstool

func ChanClose[T any](c chan T) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	close(c)
}
