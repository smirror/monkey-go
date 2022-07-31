package main

import (
	"fmt"
	"monkey-go/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Monkey programing langage!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
	fmt.Printf("Goodbye %s!\n", user.Username)
	fmt.Printf("See you again!\n")
}
