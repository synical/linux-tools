package main

import (
	"github.com/rthornton128/goncurses"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

type netDevice struct {
	rx, tx                 int
	previousRx, previousTx int
	rbps, tbps             int
	name                   string
}

func getActiveDevices() []string {
	devices := make([]string, 0)
	dirs, _ := ioutil.ReadDir("/sys/class/net/")
	for _, d := range dirs {
		dev := d.Name()
		b, _ := ioutil.ReadFile("/sys/class/net/" + d.Name() + "/operstate")
		if strings.Contains(string(b), "up") {
			devices = append(devices, dev)
		}
	}
	return devices
}

func measureThroughput(c chan netDevice, d *netDevice) {
	for {
		d.readNetBytes()
		d.rbps = d.rx - d.previousRx
		d.tbps = d.tx - d.previousTx
		d.previousRx = d.rx
		d.previousTx = d.tx
		c <- *d
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
  # Call end on SIGINT
*/
func main() {
	stdscr, _ := goncurses.Init()
	defer goncurses.End()

	activeDevices := getActiveDevices()
	if len(activeDevices) == 0 {
		log.Fatal("No active devices found!")
	}

	c := make(chan netDevice)
	for _, deviceName := range activeDevices {
		device := &netDevice{rx: 0, tx: 0, name: deviceName}
		go measureThroughput(c, device)
	}
	for {
		select {
		case d := <-c:
			stdscr.Print("\r", d.rbps)
			stdscr.Refresh()
			stdscr.Clear()
		}
	}
}
