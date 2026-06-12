package main

import "testing"

func TestUnknownCommand(t *testing.T) {
	code := run([]string{"unknown"})
	if code != 1 {
		t.Fail()
	}
}
