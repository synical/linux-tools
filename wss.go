package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type RefMem interface {
	clearRefs()
	countRefKb()
}

type Process struct {
	Pid   string
	RefKb int64
}

func check(msg string, err error) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}

func (p *Process) clearRefs() {
	f, err := os.OpenFile("/proc/"+p.Pid+"/clear_refs", os.O_WRONLY, 0666)
	check("Could not open clear_refs", err)
	f.WriteString("1")
	check("Could not write to clear refs", err)
}

func countRefKb(pid string) {
}

func main() {
	var p = Process{}
	pidArg := flag.String("p", "", "<PID>")
	flag.Parse()
	if *pidArg == "" {
		log.Fatal("Must pass in pid with -p <pid>")
	}
	p.Pid = *pidArg
	fmt.Println(p)
	p.clearRefs()
}
