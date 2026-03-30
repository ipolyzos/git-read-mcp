package handlers

import "sync/atomic"

func setCompareTextTempParentForTest(parent string) (cleanup func()) {
	compareTextTempMu.Lock()
	prev := compareTextTempParent
	compareTextTempParent = parent
	compareTextTempMu.Unlock()
	return func() {
		compareTextTempMu.Lock()
		compareTextTempParent = prev
		compareTextTempMu.Unlock()
	}
}

func setDiffDirMaxOutputForTest(n int64) (cleanup func()) {
	prev := atomic.SwapInt64(&diffDirMaxOutputOverride, n)
	return func() { atomic.StoreInt64(&diffDirMaxOutputOverride, prev) }
}

func setDiffFilesMaxOutputForTest(n int64) (cleanup func()) {
	prev := atomic.SwapInt64(&diffFilesMaxOutputOverride, n)
	return func() { atomic.StoreInt64(&diffFilesMaxOutputOverride, prev) }
}
