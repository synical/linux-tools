package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
# include <unistd.h>
*/
import "C"

var pagesize = int64(C.sysconf(C._SC_PAGESIZE))

const MB = 1048576

/* http://www.brendangregg.com/blog/2018-01-17/measure-working-set-size.html */

type Process struct {
	Pid      string
	RefMb    int64
	RSSMb    int64
	PctOfRSS float64
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
			rkb += stringToInt(kb)
		}
	}
	p.RefMb = (rkb * 1024) / MB
}

func (p *Process) getMemStats() {
	p.countRefKb()
	p.getRSS()
	p.getPctOfRSS()
}

func (p *Process) getPctOfRSS() {
	p.PctOfRSS = (float64(p.RefMb) / float64(p.RSSMb)) * 100
}

func (p *Process) getRSS() {
	d, err := ioutil.ReadFile("/proc/" + p.Pid + "/statm")
	check(err)
	rss := strings.Split(string(d), " ")[1]
	p.RSSMb = (stringToInt(rss) * pagesize) / MB
}

func stringToInt(s string) int64 {
	i, err := strconv.ParseInt(s, 0, 64)
	check(err)
	return i
}

func main() {
	var p = Process{}
	pidArg := flag.String("p", "", "<PID>")
	intervalArg := flag.Int("i", 1, "<INTERVAL>")
	flag.Parse()
	if *pidArg == "" {
		log.Fatal("Must pass in pid with -p <pid>")
	}
	p.Pid = *pidArg
	interval := time.Duration(*intervalArg)
	p.getRSS()
	p.countRefKb()
	p.clearRefs()
	fmt.Printf("WSS\t\t%%RSS\t\tRSS\n")
	for {
		time.Sleep(time.Second * interval)
		p.getMemStats()
		fmt.Printf("%dmb\t\t%.0f%%\t\t%dmb\n", p.RefMb, p.PctOfRSS, p.RSSMb)
	}
}
