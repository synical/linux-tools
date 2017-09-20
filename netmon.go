package main

import (
	"fmt"
	"github.com/rthornton128/goncurses"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type netDevice struct {
	rx, tx                 int
	previousRx, previousTx int
	rbps, tbps             int
	name                   string
}

func generateOutput(devs []*netDevice) string {
	var output string = "\r"
	for _, dev := range devs {
		output = fmt.Sprintf("%s %d %d", dev.name, dev.rbps, dev.tbps)
	}
	return output
}

func getActiveDevices() []*netDevice {
	devices := make([]*netDevice, 0)
	dirs, _ := ioutil.ReadDir("/sys/class/net/")
	for _, d := range dirs {
		dev := &netDevice{name: d.Name()}
		b, _ := ioutil.ReadFile("/sys/class/net/" + d.Name() + "/operstate")
		if strings.Contains(string(b), "up") {
			devices = append(devices, dev)
		}
	}
	return devices
}

func measureThroughput(c chan []*netDevice, devs []*netDevice) {
	for {
		for _, d := range devs {
			d.readNetBytes()
			d.rbps = d.rx - d.previousRx
			d.tbps = d.tx - d.previousTx
			d.previousRx = d.rx
			d.previousTx = d.tx
		}
		c <- devs
		time.Sleep(time.Second * 1)
	}
}

func (d *netDevice) readNetBytes() {
	r, _ := ioutil.ReadFile("/sys/class/net/" + d.name + "/statistics/rx_bytes")
	t, _ := ioutil.ReadFile("/sys/class/net/" + d.name + "/statistics/tx_bytes")
	rs := string(r)
	ts := string(t)
	rs = strings.Replace(rs, "\n", "", -1)
	ts = strings.Replace(ts, "\n", "", -1)
	d.rx, _ = strconv.Atoi(rs)
	d.tx, _ = strconv.Atoi(ts)
}

/*
  TODO
  * Make channel take list of netDevices
*/
func main() {
	c := make(chan []*netDevice)
	done := make(chan bool)
	sigs := make(chan os.Signal)

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		for {
			s := <-sigs
			switch s {
			case syscall.SIGINT:
				done <- true
			}
		}
	}()

	activeDevices := getActiveDevices()
	if len(activeDevices) == 0 {
		log.Fatal("No active devices found!")
	}

	go measureThroughput(c, activeDevices)

	for {
		select {
		case devs := <-c:
			output := generateOutput(devs)
			stdscr.Print(output)
			stdscr.Refresh()
			stdscr.Clear()
		case <-done:
			goncurses.End()
			os.Exit(0)
		}
	}
}
