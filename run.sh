#!/bin/bash
x-terminal-emulator -T "DASHBOARD" -e "bash -c 'cd dashboard; go run main.go; exec bash'" &
sleep 2
x-terminal-emulator -T "SENSOR 1" -e "bash -c 'cd sensor1; go run main.go; exec bash'" &
x-terminal-emulator -T "SENSOR 2" -e "bash -c 'cd sensor2; go run main.go; exec bash'" &
