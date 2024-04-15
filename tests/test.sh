#!/bin/bash

scriptDir="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
workingDir=$PWD
testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
mkdir -p $scriptDir/coverage
coverageFile=$scriptDir/coverage/coverage

if [ "$#" -gt 0 ]; then
  if [ "$1" == "main" ]; then
    go test -covermode=atomic -coverprofile $coverageFile.txt -failfast -v -p 1
    go tool cover -html=$coverageFile.txt -o $coverageFile.html
    exit 0
  fi

  testFile=$1
  echo "Testing $testFile"

  if [ -d "$testFile" ]; then
    cd $testFile
    testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
    for s in $testFiles; do
      if ! go test -covermode=atomic -coverprofile $coverageFile.txt -failfast -v -p 1 $s;
        then break;
      fi;
    done
  else
    testFiles=$testFile
  fi
else
  echo "Testing all"
  gotestsum -f testname -- ./... -race -count=1 -coverprofile=$coverageFile.txt -covermode=atomic
  go tool cover -html=$coverageFile.txt -o $coverageFile.html
fi


go tool cover -html=$coverageFile.txt -o $coverageFile.html

cd $workingDir
