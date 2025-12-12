package correlator

func BranchlessGreater(a, b uint32) uint32 {
	diff := int32(a - b - 1)
	return uint32(^(diff >> 31)) & 1
}

func BranchlessGreaterEqual(a, b uint32) uint32 {
	diff := int32(a - b)
	return uint32(^(diff >> 31)) & 1
}

func BranchlessLess(a, b uint32) uint32 {
	diff := int32(b - a - 1)
	return uint32(^(diff >> 31)) & 1
}

func BranchlessLessEqual(a, b uint32) uint32 {
	diff := int32(b - a)
	return uint32(^(diff >> 31)) & 1
}

func BranchlessEqual(a, b uint32) uint32 {
	diff := a ^ b
	return uint32((^diff & (diff - 1)) >> 31)
}

func BranchlessMin(a, b uint32) uint32 {
	diff := int32(a - b)
	mask := uint32(diff >> 31)
	return (a & mask) | (b & ^mask)
}

func BranchlessMax(a, b uint32) uint32 {
	diff := int32(a - b)
	mask := uint32(diff >> 31)
	return (b & mask) | (a & ^mask)
}

func BranchlessSelect(condition uint32, ifTrue, ifFalse uint32) uint32 {
	mask := uint32(-int32(condition & 1))
	return (ifTrue & mask) | (ifFalse & ^mask)
}

func BranchlessAnd(a, b uint32) uint32 {
	return a & b
}

func BranchlessOr(a, b uint32) uint32 {
	return a | b
}

func BranchlessNot(a uint32) uint32 {
	return ^a
}

func BranchlessXor(a, b uint32) uint32 {
	return a ^ b
}
