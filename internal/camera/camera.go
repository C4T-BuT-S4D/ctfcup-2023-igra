package camera

import "github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"

const (
	WIDTH  = 640 * 2
	HEIGHT = 480 * 2
)

type Camera struct {
	*object.Object
}
