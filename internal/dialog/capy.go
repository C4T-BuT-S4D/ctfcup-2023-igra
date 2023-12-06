package dialog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/eth"
)

const (
	Init int = iota
	AddressEntered
	TokenInEntered
)

type userInput struct {
	address string
	tokenID int
}

func NewCapy(token string) *Capy {
	return &Capy{
		Token: token,
		state: Init,
	}
}

type Capy struct {
	Token       string
	dialogState State
	state       int
	userInput   userInput
}

func (c *Capy) Greeting() {
	c.dialogState.Text = fmt.Sprintf("You may pet the capybara (token: %s) on address:", c.Token)
	c.state = Init
}

func (c *Capy) Feed(text string) {
	text = strings.TrimSpace(strings.ToLower(text))
	switch c.state {
	case Init:
		c.userInput.address = text
		c.state = AddressEntered
		c.dialogState.Text += fmt.Sprintf("\n> %s", text)
		c.dialogState.Text += "\nNow enter token id (number):"
	case AddressEntered:
		c.dialogState.Text += fmt.Sprintf("\n> %s", text)

		n, err := strconv.Atoi(text)
		if err != nil {
			c.dialogState.Text += "\nToken id must be a number, try again!"
			return
		}
		c.userInput.tokenID = n
		c.state = TokenInEntered

		win, err := eth.Check(c.userInput.address, c.userInput.tokenID, c.Token)
		if err != nil {
			c.dialogState.Text += fmt.Sprintf("\nError: %v. Please try again", err)
			return
		}
		if win {
			c.dialogState.GaveItem = true
			c.dialogState.Finished = true
		} else {
			c.dialogState.Text = "Wrong token id or address, please try again.\nEnter the address:"
			c.state = Init
		}
	}
}

func (c *Capy) State() *State {
	return &c.dialogState
}

func (c *Capy) SetState(_ *State) {
	// No need.
}
