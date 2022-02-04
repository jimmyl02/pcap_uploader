#!/bin/sh
tcpdump -i ens4 -C 1000 -W 40 -w dump.pcap