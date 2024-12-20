package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsers(t *testing.T) {
	s1 := "name=bob\nage=10\n"
	// bytes to Config
	c1 := Marshal([]byte(s1))
	c2 := Config{
		"name": "bob",
		"age":  "10",
	}
	assert.Equal(t, c1, c2)
	// Config to bytes
	s2 := Unmarshal(c1)
	assert.Equal(t, s1, s2)
}
