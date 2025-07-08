#!/usr/bin/env python3

"""
This script simulates a simple serial communication between a host and a pod.
It supports two modes of operation: 'host' and 'pod':
 - Host listens for a message and responds.
 - Pod sends a message and waits for a response.
Both sides are verifying the messages they send and receive.
"""

import serial
import sys

DEVICE_POD = "/dev/ttyPipeB0"
DEVICE_HOST = "/dev/ttyPipeA0"

MSG_1 = b"HELLO\n"
MSG_2 = b"THERE\n"


def send_msg(ser, msg):
    print(f"Sending message: {msg}")
    ser.write(msg)


def recv_msg(ser, expected_msg):
    print(f"Listening for a message. Expecting: {expected_msg}")
    line = ser.readline()
    print(f"Received message: {line}")
    if expected_msg != line:
        print("Received message does not match expected one")
        sys.exit(1)


def host():
    s = serial.Serial(DEVICE_HOST, timeout=60)
    recv_msg(s, MSG_1)
    send_msg(s, MSG_2)
    print("Test successful")


def pod():
    s = serial.Serial(DEVICE_POD, timeout=10)
    send_msg(s, MSG_1)
    recv_msg(s, MSG_2)
    print("Test successful")


if len(sys.argv) == 1:
    print("Not enough args")
    sys.exit(1)

mode = sys.argv[1]
if mode == "host":
    host()
elif mode == "pod":
    pod()
else:
    print(f"Invalid arg: {mode}. Accepted args: host | pod")
    sys.exit(1)
