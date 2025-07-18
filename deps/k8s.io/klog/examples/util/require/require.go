package require

func NoError(err error) {
	if err != nil {
		panic(err)
	}
}
