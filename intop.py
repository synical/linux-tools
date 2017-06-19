#!/usr/bin/env python

import argparse
import re

from collections import defaultdict
from os import system
from time import sleep

#TODO
#
# Add arg for printing by overall sum

def diff_interrupt_sums(before, after):
    interrupt_dict_diff = defaultdict(dict)
    for k, v in before.iteritems():
        diff = max(v["sum"], after[k]["sum"]) - min(v["sum"], after[k]["sum"])
        interrupt_dict_diff[k]["name"] = v["name"]
        interrupt_dict_diff[k]["sum"] = diff
    print_top_interrupts(interrupt_dict_diff)

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
            split_line = filter(None, line.strip("\n").replace(":", "").split(" "))
            device_name = " ".join(split_line[num_cpus+1:])
            cpu_sum = sum_interrupts(num_cpus, split_line)
            if cpu_sum == 0:
                continue
            interrupt_dict[split_line[0]]["name"] = device_name
            interrupt_dict[split_line[0]]["sum"] = cpu_sum
        f.close()
    return interrupt_dict

def print_top_interrupts(interrupt_dict):
        top_interrupts = sorted(interrupt_dict, key=lambda i: interrupt_dict[i]["sum"], reverse=True)
        for i in top_interrupts:
            print "%s (%s) -> %s" % (i, interrupt_dict[i]["name"], interrupt_dict[i]["sum"])

def sum_interrupts(num_cpus, split_line):
    cpu_sum = 0
    for x in range(1, num_cpus+1):
        # If interrupt not handled by all cpus continue
        try:
            cpu_sum += int(split_line[x])
        except IndexError:
            continue
    return cpu_sum

def main():
    args = parse_args()
    system("clear")
    diff_interrupt_sums(parse_interrupts(), parse_interrupts())
    while True:
        interrupt_dict_before = parse_interrupts()
        sleep(float(args.interval))
        system("clear")
        interrupt_dict_after = parse_interrupts()
        diff_interrupt_sums(interrupt_dict_before, interrupt_dict_after)

if __name__ == '__main__':
    main()
