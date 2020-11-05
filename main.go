package main

import (
	"github.com/mattolenik/cloudflare-ddns-client/cmd"
	"github.com/mattolenik/cloudflare-ddns-client/errhandler"
)

func main() {
	err := cmd.Root.Execute()
	errhandler.Handle(err)
}
