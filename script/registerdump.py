#!/usr/bin/python3

import argparse
import re
from subprocess import run

def fetch(bus, device):
    """ Dump registers/memory of some remote I2C devices. """

    command = ["i2cdump", "-y", str(bus), str(device)]
    return run(command , capture_output=True).stdout.decode()

def format(output):
    """ Format memory output into a go array. """

    # Regex matcher to match a value from the memory dump
    matcher = re.compile(r" *?([0-9a-f]{2}) ")
    # Find all values an convert their representation to 0xXX
    values = [f"0x{match.group(1)}" for match in matcher.finditer(output)]
    # Join the values into lines with 16 entries, separated by commas
    rows = ["  " + ", ".join(values[row*16: row*16+16]) for row in range(16)]
    # Join lines into one string
    content = ",\n".join(rows)
    # Embed that thing into a Go array definition
    return "data := []byte{\n" + content + ",\n}"


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("bus", help="Bus number")
    parser.add_argument("device", help="Device ID")
    args = parser.parse_args()

    print(format(fetch(args.bus, args.device)))  # Fetch format and print
