package util

func Check(err interface{}) {
	if err != nil {
		panic(err)
	}
}
