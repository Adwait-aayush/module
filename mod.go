package module

import "crypto/rand"

const source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+-%@!#$^&*()[]{}|;:,.<>?/`~"

type Module struct {
}

func (m *Module) GenRandomString(n int) string {
	s, r := make([]rune, n), []rune(source)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x,y:=p.Uint64(),uint64(len(r))
		s[i]=r[x%y]
	}
	return string(s)
}
