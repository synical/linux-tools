package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
  TODO
  - Flag for interval (default to a second)
  - Add column including WSS as percentage of RES
*/

/* http://www.brendangregg.com/blog/2018-01-17/measure-working-set-size.html */

type Process struct {
	Pid   string
	RefKb int64
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func (p *Process) clearRefs() {
	fmt.Printf("Clearing referenced flag on all pages for pid %s\n\n", p.Pid)
	f, err := os.OpenFile("/proc/"+p.Pid+"/clear_refs", os.O_WRONLY, 0666)
	check(err)
	defer f.Close()
	f.WriteString("1")
	check(err)
}

func (p *Process) countRefKb() {
	var rkb int64
	f, err := os.Open("/proc/" + p.Pid + "/smaps")
	check(err)
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text()
		if strings.HasPrefix(l, "Referenced") {
			kb := strings.Split(l, ":")[1]
			kb = strings.Replace(kb, " ", "", -1)
			kb = strings.Replace(kb, "kB", "", -1)
			rkb += stringToKb(kb)
		}
	}
	p.RefKb = rkb
}

func stringToKb(kb string) int64 {
	i, err := strconv.ParseInt(kb, 0, 64)
	check(err)
	return i
}

func main() {
	var p = Process{}
	pidArg := flag.String("p", "", "<PID>")
	flag.Parse()
	if *pidArg == "" {
		log.Fatal("Must pass in pid with -p <pid>")
	}
	p.Pid = *pidArg
	p.countRefKb()
	p.clearRefs()
	fmt.Printf("WSS\n")
	for {
		time.Sleep(time.Second * 1)
		p.countRefKb()
		fmt.Printf("%dkb\n", p.RefKb)
	}
}
