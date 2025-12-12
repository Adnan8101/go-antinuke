package util

type Xorshift32 struct {
	state uint32
}

func NewXorshift32(seed uint32) *Xorshift32 {
	if seed == 0 {
		seed = 2463534242
	}
	return &Xorshift32{state: seed}
}

func (x *Xorshift32) Next() uint32 {
	x.state ^= x.state << 13
	x.state ^= x.state >> 17
	x.state ^= x.state << 5
	return x.state
}

func (x *Xorshift32) Hash(val uint32) uint32 {
	val ^= val << 13
	val ^= val >> 17
	val ^= val << 5
	return val
}

type Xorshift64 struct {
	state uint64
}

func NewXorshift64(seed uint64) *Xorshift64 {
	if seed == 0 {
		seed = 88172645463325252
	}
	return &Xorshift64{state: seed}
}

func (x *Xorshift64) Next() uint64 {
	x.state ^= x.state << 13
	x.state ^= x.state >> 7
	x.state ^= x.state << 17
	return x.state
}

func (x *Xorshift64) Hash(val uint64) uint64 {
	val ^= val << 13
	val ^= val >> 7
	val ^= val << 17
	return val
}

func HashU32(val uint32) uint32 {
	val ^= val << 13
	val ^= val >> 17
	val ^= val << 5
	return val
}

func HashU64(val uint64) uint64 {
	val ^= val << 13
	val ^= val >> 7
	val ^= val << 17
	return val
}

func HashIndex(val, mask uint32) uint32 {
	return HashU32(val) & mask
}

func HashIndex64(val, mask uint64) uint64 {
	return HashU64(val) & mask
}
