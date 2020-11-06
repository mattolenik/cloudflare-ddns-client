package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

// This is a helper program for generating the README.md file from a template.
// It provides a function, `run`, that runs any command and places its output into
// the template's output. With this function the readme can embed the actual help
// text of the program itself, ensuring that help text in the readme is kept up to date.
//
// Example usage for a markdown readme:
//
// ````
// {{ run "go" "run" "main.go" "-help" }}
// ````

func main() {
	err := generate()
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		os.Exit(1)
	}
}

func generate() error {
	templateFile := os.Args[1]
	templateText, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}
	tpl := template.New("readme")
	tpl.Funcs(template.FuncMap{
		"run": run,
	})
	_, err = tpl.Parse(string(templateText))
	if err != nil {
		return err
	}
	err = tpl.Execute(os.Stdout, nil)
	if err != nil {
		return err
	}
	return nil
}

func run(args ...string) string {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return string(out)
}
