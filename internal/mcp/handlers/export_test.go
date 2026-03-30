package handlers

func setCompareTextTempParentForTest(dir string) (cleanup func()) {
	compareTextTempMu.Lock()
	prev := compareTextTempParent
	compareTextTempParent = dir
	compareTextTempMu.Unlock()
	return func() {
		compareTextTempMu.Lock()
		compareTextTempParent = prev
		compareTextTempMu.Unlock()
	}
}
