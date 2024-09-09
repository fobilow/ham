package helper

import "os"

func GetEnv(name, defaultValue string) string {
	v, set := os.LookupEnv(name)
	if !set {
		return defaultValue
	}
	return v
}
