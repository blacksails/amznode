package amznode_test

import (
	"os"
	"testing"

	"github.com/blacksails/amznode"
)

func TestGetEnv(t *testing.T) {
	cases := map[string]struct {
		env        map[string]string
		envStr     string
		defaultStr string
		result     string
	}{
		"returns existing variables": {
			env:        map[string]string{"TEST": "HELLO"},
			envStr:     "TEST",
			defaultStr: "CHEESE",
			result:     "HELLO",
		},
		"returns default on non existing variables": {
			env:        map[string]string{"TEST": "HELLO"},
			envStr:     "NONEXISTANT",
			defaultStr: "CHEESE",
			result:     "CHEESE",
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			for k, v := range c.env {
				os.Setenv(k, v)
			}
			if got := amznode.GetEnv(c.envStr, c.defaultStr); got != c.result {
				t.Errorf("Expected %v, got %v", c.result, got)
			}
		})
	}
}
