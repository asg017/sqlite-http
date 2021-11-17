#!/bin/bash

#set -e

for TEST in tests/*.sql; do 
  echo "Running test $TEST"
  ~/Downloads/sqlite-tools-osx-x86-3360000/sqlite3 ':memory:' ".read $TEST"
  retVal=$?
  if [ $retVal -ne 0 ]; then
      echo "$TEST failed with code=$retVal"
      exit 1
  fi
done

echo "Tests successful!"