#!/usr/bin/env python

import re

from collections import defaultdict
from os import system
from time import sleep

#TODO
#
# Order by greatest delta and not sum
# Make get_top_interrupts function

def get_cpus(cpu_line):
    cpu_regex = r"CPU[0-9]+"
    return re.findall(cpu_regex, cpu_line)

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
    system("clear")
    interrupt_dict = defaultdict(dict)
    while True:
        with open("/proc/interrupts") as f:
                for i, line in enumerate(f.readlines()):
                    if i == 0:
                        cpus = get_cpus(line)
                        num_cpus = len(cpus)
                        continue
                    split_line = filter(None, line.strip("\n").replace(":", "").split(" "))
                    device_name = " ".join(split_line[num_cpus+1:])
                    cpu_sum = sum_interrupts(num_cpus, split_line)
                    interrupt_dict[split_line[0]]["name"] = device_name
                    interrupt_dict[split_line[0]]["sum"] = cpu_sum
                top_interrupts = sorted(interrupt_dict, key=lambda i: interrupt_dict[i]["sum"], reverse=True)
                for i in top_interrupts:
                    print "%s (%s) -> %s" % (i, interrupt_dict[i]["name"], interrupt_dict[i]["sum"])
                f.close()
                sleep(2)
                system("clear")

if __name__ == '__main__':
    main()
