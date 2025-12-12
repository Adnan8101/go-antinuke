package logging

type LZ4BufferPool struct {
	buffers chan []byte
	size    int
}

func NewLZ4BufferPool(poolSize, bufferSize int) *LZ4BufferPool {
	pool := &LZ4BufferPool{
		buffers: make(chan []byte, poolSize),
		size:    bufferSize,
	}

	for i := 0; i < poolSize; i++ {
		pool.buffers <- make([]byte, bufferSize)
	}

	return pool
}

func (lbp *LZ4BufferPool) Get() []byte {
	select {
	case buf := <-lbp.buffers:
		return buf
	default:
		return make([]byte, lbp.size)
	}
}

func (lbp *LZ4BufferPool) Put(buf []byte) {
	if len(buf) != lbp.size {
		return
	}

	select {
	case lbp.buffers <- buf:
	default:
	}
}
