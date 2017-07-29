#!/usr/bin/env python

import curses
import argparse
import re

from collections import defaultdict
from curses import wrapper
from os import system
from time import sleep

# TODO
# Top like UI with columns
# Fix output when number of interrupts is huge
# Add filter by interrupt

def diff_interrupt_sums(before, after):
    interrupt_dict_diff = defaultdict(dict)
    for k, v in before.iteritems():
        diff = max(v["sum"], after[k]["sum"]) - min(v["sum"], after[k]["sum"])
        if diff == 0:
            continue
        interrupt_dict_diff[k]["name"] = v["name"]
        interrupt_dict_diff[k]["sum"] = diff
    return interrupt_dict_diff

def get_cpus(cpu_line):
    cpu_regex = r"CPU[0-9]+"
    return re.findall(cpu_regex, cpu_line)

def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", "--interval", action="store", dest="interval", default=5, metavar="<interval>", help="Sampling interval")
    return parser.parse_args()

def parse_interrupts():
    interrupt_dict = defaultdict(dict)
    with open("/proc/interrupts") as f:
        for i, line in enumerate(f.readlines()):
            if i == 0:
                cpus = get_cpus(line)
                num_cpus = len(cpus)
                continue
            split_line = filter(None, line.strip().replace(":", "").split())
            device_name = " ".join(split_line[num_cpus+1:])
            cpu_sum = sum_interrupts(num_cpus, split_line)
            if cpu_sum == 0:
                continue
            interrupt_dict[split_line[0]]["name"] = device_name
            interrupt_dict[split_line[0]]["sum"] = cpu_sum
        f.close()
    return interrupt_dict

def print_top_interrupts(window, interrupt_dict):
        top_interrupts = sorted(interrupt_dict, key=lambda i: interrupt_dict[i]["sum"], reverse=True)
        total_interrupts = sum([v["sum"] for k,v in interrupt_dict.iteritems()])
        window.addstr("Total -> %s\n" % (total_interrupts))
        for i in top_interrupts:
            window.addstr("%s (%s) -> %s\n" % (i, interrupt_dict[i]["name"], interrupt_dict[i]["sum"]))

def sum_interrupts(num_cpus, split_line):
    cpu_sum = 0
    for x in range(1, num_cpus+1):
        # If interrupt not handled by all cpus continue
        try:
            cpu_sum += int(split_line[x])
        except IndexError:
            continue
    return cpu_sum

def top_interrupt_loop(window, interval):
    while True:
        interrupt_dict_before = parse_interrupts()
        window.refresh()
        sleep(interval)
        window.erase()
        interrupt_dict_after = parse_interrupts()
        interrupts_diff = diff_interrupt_sums(interrupt_dict_before, interrupt_dict_after)
        window.addstr("Interrupts per %s seconds\n\n" % (interval))
        print_top_interrupts(window, interrupts_diff)

def main(window):
    args = parse_args()
    curses.curs_set(0)
    window.addstr("Top interrupts since boot\n\n")
    print_top_interrupts(window, parse_interrupts())
    top_interrupt_loop(window, float(args.interval))

if __name__ == '__main__':
    wrapper(main)
