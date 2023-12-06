package dialog

import "strings"

const MODULO int = 31337

func get_mod(n, m int) int {
	return ((n % m) + m) % m
}

func poly_mul(poly1, poly2 []int) []int {
	res := make([]int, len(poly1)+len(poly2)-1)
	for p1, c1 := range poly1 {
		for p2, c2 := range poly2 {
			res[p1+p2] = get_mod(res[p1+p2]+c1*c2, MODULO)
		}
	}

	return res
}

func check(flag string) bool {
	poly := []int{1}
	for i, c := range []byte(flag) {
		poly = poly_mul(poly, []int{get_mod(int(c)+(i<<8), MODULO), 1})
	}

	target := []int{815, 4966, 27829, 14420, 17439, 22697, 12685, 30092, 7096, 30329, 870, 22629, 20414, 23532, 15967, 448, 1}

	if len(target) != len(poly) {
		return false
	}

	for i, c := range poly {
		if c != target[i] {
			return false
		}
	}
	return true
}

func NewWiseTree() *WiseTree {
	return &WiseTree{
		state: State{},
	}
}

type WiseTree struct {
	state State
}

func (w *WiseTree) Greeting() {
	w.state.Text = "I am a wise tree. What do you want to know?"
}

func (w *WiseTree) Feed(text string) {
	prefix := "My honest reaction to that information:"
	text = strings.ToLower(text)
	if check(text) {
		w.state.Text = prefix + " wise"
		w.state.GaveItem = true
		w.state.Finished = true
		return
	}
	w.state.Text = prefix + " bruh"
	return
}

func (w *WiseTree) State() *State {
	return &w.state
}

func (w *WiseTree) SetState(s *State) {
	// No need.
}
