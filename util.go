package zen

// assert c is true, else panic with msg
func assert(c bool, msg string) {
	if !c {
		panic(msg)
	}
}
