package main

import (
    "fmt"
    "os"
    "os/user"
    "monkey/repl"
)

func main() {
    user, err := user.Current()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Hello %s, The Monkey programming language is here!", user.Username)
    fmt.Printf(" Type a command...")

    repl.Start(os.Stdin, os.Stdout)
}
