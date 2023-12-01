package dialog

type State struct {
	Finished bool
	GaveItem bool
	Text     string
}
type Dialog interface {
	Greeting()
	Feed(text string)
	State() State
}
