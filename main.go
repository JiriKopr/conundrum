package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type Node struct {
	subnodes map[string]Node
}

func (node Node) InsertWord(word string) {
	firstChar := word[0:1]

	foundNode, exists := node.subnodes[firstChar]

	if !exists {
		foundNode = Node{subnodes: make(map[string]Node)}
		node.subnodes[firstChar] = foundNode
	}

	if len(word) > 1 {
		// There is another character
		foundNode.InsertWord(word[1:])
	} else if len(word) == 1 {
		// No more characters
		foundNode.subnodes[""] = Node{subnodes: make(map[string]Node)}
	}
}

func (node Node) FindWord(word string) bool {

	firstChar := word[0:1]

	foundNode, exists := node.subnodes[firstChar]

	if exists {
		if len(word) == 1 {
			_, exists = foundNode.subnodes[""]
			return exists
		}

		return foundNode.FindWord(word[1:])
	}

	return false
}

func join(ins []rune, c rune) (result []string) {
	for i := 0; i <= len(ins); i++ {
		result = append(result, string(ins[:i])+string(c)+string(ins[i:]))
	}
	return
}

func permutations(testStr string) []string {
	var n func(testStr []rune, p []string) []string
	n = func(testStr []rune, p []string) []string {
		if len(testStr) == 0 {
			return p
		} else {
			result := []string{}
			for _, e := range p {
				result = append(result, join([]rune(e), testStr[0])...)
			}
			return n(testStr[1:], result)
		}
	}

	output := []rune(testStr)

	perms := n(output[1:], []string{string(output[0])})

	return unique(perms)
}

func worker(nodeWord Node, results chan<- string, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for lookupWord := range jobs {
		if nodeWord.FindWord(lookupWord) {
			results <- lookupWord
		}
	}
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func loadWords() Node {
	word_bytes, err := ioutil.ReadFile("words.txt")
	if err != nil {
		fmt.Println("Error while reading file", err)
		os.Exit(1)
	}

	words := strings.Split(string(word_bytes), "\n")

	root := Node{subnodes: make(map[string]Node)}

	for _, word := range words {
		word := strings.TrimSpace(word)
		root.InsertWord(word)
	}

	return root

}

func loadLookupWord() string {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("No lookup word supplied")
		os.Exit(1)
	}

	lookupWord := string(args[0])
	return strings.TrimSpace(lookupWord)
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

func main() {
	defer elapsed("page")()
	lookupWord := loadLookupWord()

	rootWord := loadWords()

	wordPerm := permutations(lookupWord)

	results := make(chan string)
	jobs := make(chan string, len(wordPerm))
	numberOfWorkers := 5
	wg := &sync.WaitGroup{}

	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go worker(rootWord, results, jobs, wg)
	}

	for _, perm := range wordPerm {
		jobs <- perm
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		fmt.Println(res)
	}
}
