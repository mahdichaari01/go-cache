package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	goCache "github.com/mahdichaari01/go-cache/cache"
)

func main() {

	runInteractiveMode()

}

func runInteractiveMode() {
	scanner := bufio.NewScanner(os.Stdin)
	var cache *goCache.LruCache
	fmt.Println("Welcome to the interactive demo of go-cache")

	for {
		fmt.Print("\nEnter Cache Size: ")
		scanner.Scan()
		cmd := scanner.Text()

		// check if cmd is a number
		i, err := strconv.Atoi(cmd)

		if err != nil {
			fmt.Printf("%s is not a number!", cmd)
			continue
		}
		// attempt cache creation
		if cache, err = goCache.NewCache(i); err == nil {
			break
		}

		fmt.Printf("Error Creating Cache: %s", err)
	}

	for {
		fmt.Print("\nEnter command ((s)et/(g)et/(d)elete/(q)uit): ")
		scanner.Scan()
		cmd := scanner.Text()

		switch cmd {
		case "q":
			return
		case "s":
			fmt.Print("Enter key: ")
			scanner.Scan()
			key := scanner.Text()
			fmt.Print("Enter value: ")
			scanner.Scan()
			value := scanner.Text()
			cache.Set(key, value)
			fmt.Println("Value set successfully")
		case "g":
			fmt.Print("Enter key: ")
			scanner.Scan()
			key := scanner.Text()
			if value, exists := cache.Get(key); exists {
				fmt.Printf("Value: %s\n", value)
			} else {
				fmt.Println("Key not found")
			}
		case "d":
			fmt.Print("Enter key: ")
			scanner.Scan()
			key := scanner.Text()
			if cache.Delete(key) {
				fmt.Println("Key deleted successfully")
			} else {
				fmt.Println("Key not found")
			}
		default:
			fmt.Println("Unknown command")
		}
	}
}
