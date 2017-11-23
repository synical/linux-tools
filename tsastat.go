package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

/*
Print time spent in each of the six thread states
  - Start for all threads of a process
  - Diff per interval with interval arg passed

I think the point of this in the Brendan Gregg book
is to demonstrate that this is quite hard in Linux
*/

func getThreadState(pid string) string {
	status, err := ioutil.ReadFile("/proc/" + pid + "/status")
	if err != nil {
		log.Fatal("Failed to read status file: ", err)
	}
	stateLine := strings.Split(string(status), "\n")[2]
	state := string(strings.Split(stateLine, ":")[1][1])
	return state
}

func main() {
	tsMap := make(map[int]string)
	dirs, err := ioutil.ReadDir("/proc/")
	if err != nil {
		log.Fatal("Failed to read /proc: ", err)
	}

	for _, dir := range dirs {
		name := dir.Name()
		if pid, err := strconv.Atoi(name); err == nil {
			tsMap[pid] = getThreadState(name)
		}
	}

	for pid, state := range tsMap {
		fmt.Printf("%d: %s\n", pid, state)
	}

	return
}
