// package conf contains functions to parse a Config
package conf

import (
	"bufio"
	"bytes"
	"strings"
)

type Config map[string]string

// Marshal parses the bytes into a Config and returns it.
func Marshal(b []byte) Config {
	c := make(Config)
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		parts := strings.Split(sc.Text(), "=")
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			c[k] = v
		}
	}
	return c
}

// Unmarshal unmarshals the Config into a string and returns it.
func Unmarshal(c Config) string {
	var b strings.Builder
	for k, v := range c {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString("\n")
	}
	return b.String()
}
