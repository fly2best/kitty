#!/bin/bash

go run kittyd.go 2>&1 | tee -a kittyd.log
