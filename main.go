package main

//go:generate go run ./.tools/version_gen.go drish

import (
	_ "embed"
	"github.com/davidalpert/selanger/internal/cmd"
)

func main() {
	cmd.Execute()
}
