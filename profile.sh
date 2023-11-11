#!/bin/bash
go test -bench ^BenchmarkPlan$ -benchtime=3s -benchmem -memprofile mem.pprof -cpuprofile cpu.pprof
go tool pprof -http=:8080 cpu.pprof