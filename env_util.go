package amznode

import "os"

// GetEnv returns the value of the environment variable or optionally a
// default value.
func GetEnv(envStr string, defaultStr string) string {
	str := os.Getenv(envStr)
	if str == "" {
		return defaultStr
	}
	return str
}
