package main

import (
	"io/ioutil"
	"log"
	"fmt"
	"strings"
	"os"
	"strconv"
)

// contains determines if a given out entry matches any answer entry.
func contains(slice []string, findMe string) bool {
	findMeSplit := strings.Split(findMe, "\n")
	for _, element := range slice {
		elementSplit := strings.Split(element, "\n")
		if len(elementSplit) != 7 {
			return false
		}
		var counter int
		for i, _ := range findMeSplit {
			if strings.Trim(elementSplit[i], " ") == strings.Trim(findMeSplit[i], " ") {
				counter++
			}
		}
		if counter == 7 {
			return true
		}
	}
	return false
}

// main makes sure that the produced output matches the answer key.
func main() {
	fileToCheck := os.Args[1]
	out, err := ioutil.ReadFile(fileToCheck)
	if err != nil {
		log.Fatal(err)
	}

	fileForChecking := os.Args[2]
	ans, err := ioutil.ReadFile(fileForChecking)
	if err != nil {
		log.Fatal(err)
	}

	bestInGenreOut := strings.Split(string(out), "\n\n")[1:]
	bestInGenreOut = bestInGenreOut[:len(bestInGenreOut)-1]
	bestInGenreAns := strings.Split(string(ans), "\n\n")[1:]
	bestInGenreAns = bestInGenreAns[:len(bestInGenreAns)-1]

	if len(bestInGenreOut) != len(bestInGenreAns) {
		fmt.Println("FAIL!: Answer and output are different lengths. Output:", strconv.Itoa(len(bestInGenreOut)), ", Answer:", strconv.Itoa(len(bestInGenreAns)))
		return
	} else {
		for _, entry := range bestInGenreAns {
			if !contains(bestInGenreOut, entry) {
				fmt.Println("FAIL!:{", entry, "}\nNot in answer.")
				return
			}
		}
		fmt.Println("Ok!")
	}
}