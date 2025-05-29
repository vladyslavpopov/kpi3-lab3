#!/bin/bash

x=0.5
y=0.5

curl -X POST http://localhost:17000 -d "figure $x $y"
curl -X POST http://localhost:17000 -d "update"
sleep 1

for i in {1..20}; do
    curl -X POST http://localhost:17000 -d "move 0.01 0.01"
    curl -X POST http://localhost:17000 -d "update"
    sleep 1
done