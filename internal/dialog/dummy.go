package dialog

import "fmt"

func NewDummy(s string) Dialog {
	return &dummyDialog{greet: s}
}

type dummyDialog struct {
	s     State
	greet string
}

func (d *dummyDialog) Greeting() {
	d.s.Text = d.greet
}

func (d *dummyDialog) Feed(text string) {
	d.s.Text += fmt.Sprintf("\nYou said: %s", text)
}

func (d *dummyDialog) State() State {
	return d.s
}
