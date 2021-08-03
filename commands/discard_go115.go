// +build !go1.16

package commands

var Discard = discard{}

type discard struct {
}

func (d discard) Write(b []byte) (int, error) {
	return len(b), nil
}
