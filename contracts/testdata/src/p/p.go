package p

import "errors"

func F1(in int) error {
	if in == 0 { // want "contract: pre-condition failed"
		return errors.New("must not be zero")
	}
	return nil
}
