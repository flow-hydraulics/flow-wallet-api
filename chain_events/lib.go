package chain_events

func min(x, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}
