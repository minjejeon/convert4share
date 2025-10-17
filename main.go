package main

import (
	_ "embed"

	"github.com/minjejeon/convert4share/cmd"
)

//go:embed config.example.yaml
var configTemplate []byte

func init() {
	cmd.ConfigTemplate = configTemplate
}

func main() {
	cmd.Execute()
}
