package build

func Poll() {
	build, ok := Pop()
	if ok {
		build.Start()
	}
}
