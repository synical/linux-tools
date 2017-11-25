package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

/*
Print time spent in each of the six thread states
  - Start for all threads of a process
  - Diff per interval with interval arg passed

I think the point of this in the Brendan Gregg book
is to demonstrate that this is quite hard in Linux
*/

func getThreadState(taskPath string) string {
	status, err := ioutil.ReadFile(taskPath + "/status")
	if err != nil {
		log.Fatal("Failed to read status file: ", err)
	}
	stateLine := strings.Split(string(status), "\n")[2]
	state := string(strings.Split(stateLine, ":")[1][1])
	return state
}

func main() {
	pidArg := flag.String("p", "", "")
	flag.Parse()
	pid := *pidArg

	if _, err := os.Stat("/proc/" + pid); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("PID '" + pid + "' does not exist")
		}
	}

	taskPath := "/proc/" + pid + "/task/"
	tsMap := make(map[string]string)

	dirs, err := ioutil.ReadDir(taskPath)
	if err != nil {
		log.Fatal("Failed to read task dir: ", err)
	}

	for _, thread := range dirs {
		name := thread.Name()
		tsMap[name] = getThreadState(taskPath + name)
	}

	for thread, state := range tsMap {
		fmt.Printf("%s: %s\n", thread, state)
	}

	return
}
