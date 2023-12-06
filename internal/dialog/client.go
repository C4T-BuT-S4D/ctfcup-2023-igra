package dialog

func NewClientDialog() Dialog {
	return &ClientDialog{
		s: State{},
	}
}

type ClientDialog struct {
	s State
}

func (c *ClientDialog) Greeting() {
	// No need to greet.
}

func (c *ClientDialog) Feed(_ string) {
	// No need to feed.
	return
}

func (c *ClientDialog) State() *State {
	return &c.s
}

func (c *ClientDialog) SetState(s *State) {
	c.s = *s
}
