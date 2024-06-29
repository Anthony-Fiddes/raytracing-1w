#!/bin/sh
go test -bench . -benchmem -memprofile memprofile -cpuprofile cpuprofile -benchtime 5x
