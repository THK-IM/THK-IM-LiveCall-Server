package common

func SafeOnceWrite(ch chan bool, value bool) (isClosed bool) {
	defer func() {
		if recover() != nil {
			// 在panic后，会进入这个if中，然后将返回值改变为true，代表通道是关闭的
			isClosed = true
		}
	}()

	ch <- value // 如果通道关闭，这里会报panic
	close(ch)
	return false // 如果可以正常写入，这里会返回false，代表通道不是关闭的
}
