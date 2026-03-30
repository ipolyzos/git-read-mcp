package handlers

import "sync/atomic"

func setDiffDirMaxOutputForTest(n int64) (cleanup func()) {
	prev := atomic.SwapInt64(&diffDirMaxOutputOverride, n)
	return func() { atomic.StoreInt64(&diffDirMaxOutputOverride, prev) }
}
