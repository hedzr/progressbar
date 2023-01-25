package cursor

// var out = os.Stdout

func Write(b []byte) (n int, e error) {
	return safeWrite(b)
}
