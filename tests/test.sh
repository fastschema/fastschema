#!/bin/bash

scriptDir="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
workingDir=$PWD
testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
mkdir -p $scriptDir/coverage
coverageFile=$scriptDir/coverage/coverage

if [ "$#" -gt 0 ]; then
  if [ "$1" == "main" ]; then
    go test -covermode=atomic -coverprofile $coverageFile.txt -failfast -race -v -p 1
    go tool cover -html=$coverageFile.txt -o $coverageFile.html
    exit 0
  fi

  testFile=$1

  if [ -d "$testFile" ]; then
    echo "Testing dir $testFile"
    cd $testFile
    testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
    for s in $testFiles; do
      if ! go test -covermode=atomic -coverprofile $coverageFile.txt -failfast -race -v -p 1 $s;
        then break;
      fi;
    done
  else
    echo "Testing file $testFile"
    fileDir=$(dirname $testFile)
    cd $fileDir
    go test -covermode=atomic -coverprofile $coverageFile.txt -failfast -race -v -p 1;
  fi
else
  echo "Testing all"
  gotestsum -f testname -- ./... -failfast -race -count=1 -v -coverprofile=$coverageFile.txt -covermode=atomic
  go tool cover -html=$coverageFile.txt -o $coverageFile.html
fi

go tool cover -html=$coverageFile.txt -o $coverageFile.html

cd $workingDir
