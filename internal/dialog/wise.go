package dialog

const MODULO int = 31337

func getMod(n, m int) int {
	return ((n % m) + m) % m
}

func polyMul(poly1, poly2 []int) []int {
	res := make([]int, len(poly1)+len(poly2)-1)
	for p1, c1 := range poly1 {
		for p2, c2 := range poly2 {
			res[p1+p2] = getMod(res[p1+p2]+c1*c2, MODULO)
		}
	}

	return res
}

func check(flag string) bool {
	poly := []int{1}
	for i, c := range []byte(flag) {
		poly = polyMul(poly, []int{getMod(-int(c)-(i<<8), MODULO), 1})
	}

	target := []int{1837, 14688, 26533, 18612, 26274, 9840, 11452, 19408, 19989, 9381, 9839, 14074, 14090, 845, 8078, 31049, 1}

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
	if check(text) {
		w.state.Text = prefix + " wise"
		w.state.GaveItem = true
		w.state.Finished = true
		return
	}
	w.state.Text = prefix + " bruh"
}

func (w *WiseTree) State() *State {
	return &w.state
}

func (w *WiseTree) SetState(_ *State) {
	// No need.
}
