package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// https://golang.org/cmd/cgo/#hdr-Go_references_to_C

/*
#include <unistd.h>
*/
import "C"

/*
Print time spent in each of the six thread states
  - Start for all threads of a process
  - utime and stime for user and system time!
  - Diff per interval with interval arg passed

I think the point of this in the Brendan Gregg book
is to demonstrate that this is quite hard in Linux
*/

var sc_clk_tck C.long = C.sysconf(C._SC_CLK_TCK)

/*
https://stackoverflow.com/questions/16726779/how-do-i-get-the-total-cpu-usage-of-an-application-from-proc-pid-stat
*/

func getThreadStateInfo(tid string) map[string]float64 {
	m := make(map[string]float64)
	data := readFileWithError("/proc/" + tid + "/task/" + tid + "/stat")
	stat := strings.Split(string(data), " ")
	getCpuUsage(stat, m)
	return m
}

func getCpuUsage(stat []string, m map[string]float64) {
	uptime := getUptime()
	utime, _ := strconv.ParseFloat(stat[13], 64)
	stime, _ := strconv.ParseFloat(stat[14], 64)
	total_time := utime + stime
	start_time, _ := strconv.ParseFloat(stat[21], 64)
	seconds := uptime - (start_time / float64(sc_clk_tck))
	user_usage := ((utime / float64(sc_clk_tck)) / seconds) * 100
	system_usage := ((stime / float64(sc_clk_tck)) / seconds) * 100
	total_usage := ((total_time / float64(sc_clk_tck)) / seconds) * 100
	m["user_usage"] = user_usage
	m["system_usage"] = system_usage
	m["total_usage"] = total_usage
}

func getUptime() float64 {
	u := readFileWithError("/proc/uptime")
	us := strings.Split(string(u), " ")[0]
	uf, _ := strconv.ParseFloat(us, 64)
	return uf
}

func readFileWithError(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read file (%s): (%s)", path, err))
	}
	return data
}

func threadStateLoop(taskPath string, interval time.Duration) {
	tsMap := make(map[string]map[string]float64)
	for {
		dirs, err := ioutil.ReadDir(taskPath)
		if err != nil {
			log.Fatal("Failed to read task dir: ", err)
		}

		for _, thread := range dirs {
			name := thread.Name()
			tsMap[name] = getThreadStateInfo(name)
		}

		fmt.Printf("TID\tUSR\tSYS\tCPU\n")
		for thread, m := range tsMap {
			fmt.Printf("%s\t%.2f\t%.2f\t%.2f\n", thread, m["user_usage"], m["system_usage"], m["total_usage"])
		}
		time.Sleep(time.Second * interval)
	}
}

func main() {
	pidArg := flag.String("p", "", "")
	flag.Parse()
	pid := *pidArg

	if pid == "" {
		log.Fatal("Must pass in pid with -p <pid>")
	}

	if _, err := os.Stat("/proc/" + pid); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("PID '" + pid + "' does not exist")
		}
	}

	taskPath := "/proc/" + pid + "/task/"
	threadStateLoop(taskPath, time.Duration(1))

	return
}
