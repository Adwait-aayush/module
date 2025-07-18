package module

import "testing"

func TestMod_GenRandomString(t *testing.T) {
	var testmod Module

	s := testmod.GenRandomString(10)
	if len(s) != 10 {
		t.Error("Wrong length of random string returned")
	}
}
