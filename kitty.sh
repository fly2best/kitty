#!/bin/bash

go run kitty.go 2>&1 | tee -a kitty.log
