package util

// Assert is used to enforce a condition is true.
func Assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}
