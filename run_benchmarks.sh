#!/bin/sh
go test -bench . -benchmem -memprofile memprofile -benchtime 5x
