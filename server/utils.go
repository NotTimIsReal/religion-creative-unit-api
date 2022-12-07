package server

import (
	"os"
	"strings"
)

func readEnv() {
	bytes, err := os.ReadFile(".env")
	if err != nil {
		panic(err)
	}
	out := string(bytes)
	if len(out) == 0 {
		return
	}
	envs := strings.Split(out, "\n")
	for _, env := range envs {
		envPair := strings.Split(env, "=")
		envPair[0] = strings.TrimSpace(envPair[0])
		envPair[1] = strings.TrimSpace(envPair[1])
		//remove any " or '
		envPair[1] = strings.ReplaceAll(envPair[1], "\"", "")
		envPair[1] = strings.ReplaceAll(envPair[1], "'", "")
		os.Setenv(envPair[0], envPair[1])
	}

}
