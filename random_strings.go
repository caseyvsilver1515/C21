package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func main() {
	random_strings()

}

func random_strings() {
	rand.Seed(time.Now().Unix())

	//Only lowercase
	charSet := "abcdedfghijklmnopqrst"
	var output strings.Builder
	length := 10
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	fmt.Println(output.String())
	output.Reset()

	//Lowercase and Uppercase Both
	charSet = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	length = 20
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	fmt.Println(output.String())
}
