package main

import "github.com/wherd/wrk/cmd"
import "golang.design/x/clipboard"

func main() {
	if err := clipboard.Init(); err != nil {
		panic(err)
	}

	cmd.Execute()
}
