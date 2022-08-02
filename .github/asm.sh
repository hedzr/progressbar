#!/bin/bash

go build -o ./bin/small ./examples/small
go tool objdump -S ./bin/small >./bin/small.S
