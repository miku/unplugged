package main

import (
	"fmt"
	"time"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
)

func main() {
	t := 1 * time.Second
	for range 3 {
		fmt.Printf("user query + messages/memory + tools => %s%sllm%s\n", Cyan, Bold, Reset)
		time.Sleep(t)
		fmt.Printf("response or tool call                <= %s%sllm%s\n", Cyan, Bold, Reset)
		time.Sleep(t)
		fmt.Printf("            %s.... call tool ....%s\n", Yellow, Reset)
		fmt.Printf("add response to %s%scontext%s\n", Green, Bold, Reset)
		time.Sleep(t)
		fmt.Println("...")
		time.Sleep(t)
	}
	fmt.Println("excellent, the problem is now solved in a more efficient way.")
	fmt.Println("should we add a test case now or work on some new features?")
}
