package p

func F1(in int) {
	if in == 0 {
		panic("must not be zero")
	}
}

func F2(in int) {
	F1(in)
	F1(0) // want "contract violated: must not be zero"
}
