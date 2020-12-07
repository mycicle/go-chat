package main

import (
	"fmt"
	"log"
	"example.com/greetings"
	"rsc.io/quote"
)

func main() {
	// fmt.Println("Hello, World!")
	fmt.Println(quote.Go())

	log.SetPrefix("greetings: ")
	log.SetFlags(0)

	names := []string {"Michael", "James", "Andrew"}

	message, err := greetings.Hellos(names)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(message)
}