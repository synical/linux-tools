#!/usr/bin/env python

import re

from collections import defaultdict

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
    interrupt_dict = {}
    with open("/proc/interrupts") as f:
        for i, line in enumerate(f.readlines()):
            if i == 0:
                cpus = get_cpus(line)
                num_cpus = len(cpus)
                continue
            split_line = filter(None, line.strip("\n").replace(":", "").split(" "))
            cpu_sum = sum_interrupts(num_cpus, split_line)
            interrupt_dict[split_line[0]] = cpu_sum
        top_interrupts = sorted(interrupt_dict, key=interrupt_dict.get, reverse=True)
        for i in top_interrupts:
            print "[%s] -> %s" % ()

if __name__ == '__main__':
    main()
