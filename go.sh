#!/bin/bash

export VEDING_MACHINE_PSQL_DSN=postgresql://localhost:5432/veding_machine
export VEDING_MACHINE_PORT=:8080

go run main.go