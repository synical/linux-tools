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

var sc_clk_tck float64 = float64(C.sysconf(C._SC_CLK_TCK))

func calculateCpuTime(before map[string]map[string]interface{}, after map[string]map[string]interface{}, interval float64) {
	for name, _ := range after {
		before[name]["user_usage"] = (((after[name]["user_usage"].(float64) - before[name]["user_usage"].(float64)) / sc_clk_tck) / interval) * 100
		before[name]["system_usage"] = (((after[name]["system_usage"].(float64) - before[name]["system_usage"].(float64)) / sc_clk_tck) / interval) * 100
		before[name]["total_usage"] = (((after[name]["total_usage"].(float64) - before[name]["total_usage"].(float64)) / sc_clk_tck) / interval) * 100
	}
}

func getThreadStateInfo(tid string) map[string]interface{} {
	m := make(map[string]interface{})
	data := readFileWithError("/proc/" + tid + "/task/" + tid + "/stat")
	stat := strings.Split(string(data), " ")
	getCpuUsage(stat, m)
	getProcessor(stat, m)
	getTaskState(stat, m)
	return m
}

func getCpuUsage(stat []string, m map[string]interface{}) {
	utime, _ := strconv.ParseFloat(stat[13], 64)
	stime, _ := strconv.ParseFloat(stat[14], 64)
	totalTime := utime + stime
	m["user_usage"] = utime
	m["system_usage"] = stime
	m["total_usage"] = totalTime
}

func getProcessor(stat []string, m map[string]interface{}) {
	p, _ := strconv.ParseFloat(stat[38], 64)
	m["processor"] = p
}

func getTaskState(stat []string, m map[string]interface{}) {
	m["state"] = stat[2]
}

func readFileWithError(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read file (%s): (%s)", path, err))
	}
	return data
}

func threadStateLoop(taskPath string, interval time.Duration) {
	tsMap := make(map[string]map[string]interface{})
	tsMapInterval := make(map[string]map[string]interface{})
	for {
		dirs, err := ioutil.ReadDir(taskPath)
		if err != nil {
			log.Fatal("Failed to read task dir: ", err)
		}

		for _, thread := range dirs {
			name := thread.Name()
			tsMap[name] = getThreadStateInfo(name)
		}
		time.Sleep(time.Second * interval)
		for _, thread := range dirs {
			name := thread.Name()
			tsMapInterval[name] = getThreadStateInfo(name)
		}
		calculateCpuTime(tsMap, tsMapInterval, float64(interval))

		fmt.Printf("PID\tSTA\tCPU\tUSR\tSYS\tTOT\n")
		for thread, m := range tsMap {
			fmt.Printf("%s\t%s\t%.0f\t%.2f\t%.2f\t%.2f\n",
				thread,
				m["state"],
				m["processor"],
				m["user_usage"],
				m["system_usage"],
				m["total_usage"])
		}
	}
}

func main() {
	pidArg := flag.String("p", "", "<PID>")
	intervalArg := flag.Int("i", 1, "<INTERVAL>")
	flag.Parse()

	pid := *pidArg
	interval := *intervalArg

	if pid == "" {
		log.Fatal("Must pass in pid with -p <pid>")
	}

	if _, err := os.Stat("/proc/" + pid); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("PID '" + pid + "' does not exist")
		}
	}

	taskPath := "/proc/" + pid + "/task/"
	threadStateLoop(taskPath, time.Duration(interval))
	return
}
