// Package hostid is the machine-identity rule stamped on every Record and the
// control session. Change this package when the rule changes; the bus, the
// runtime kit, and the wire only carry the resulting string.
package hostid

import (
	"errors"
	"os"
)

// Resolve returns the machine identity. A non-empty override wins; otherwise
// the kernel hostname. Clone collisions are an operator problem: pass override.
func Resolve(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	h, err := os.Hostname()
	if err != nil {
		return "", err
	}
	if h == "" {
		return "", errors.New("hostid: empty hostname")
	}
	return h, nil
}
