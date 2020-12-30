package locator

var Location struct {
	Countries []struct {
		Name   string
		Cities []struct {
			Name string
		}
	}
}
