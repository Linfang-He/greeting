#!/bin/bash

names="Alfa Bravo Charlie Delta Echo Foxtrot Golf"

for name in ${names}
do
  echo "Send Hello ${name}" to the host
  echo -n "Hello ${name}" | nc -w0 localhost 8080
done