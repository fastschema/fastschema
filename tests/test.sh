#!/bin/bash

scriptDir="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
workingDir=$PWD
testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
mkdir -p $scriptDir/coverage
coverageFile=$scriptDir/coverage/coverage

if [ "$#" -gt 0 ]; then
  testFile=$1
  echo "Testing $testFile"

  if [ -d "$testFile" ]; then
    cd $testFile
    testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
  else
    testFiles=$testFile
  fi
else
  echo "Testing all"
  testFiles=$(go list ./... | grep -v /.vscode | grep -v /tests/data)
fi

for s in $testFiles; do
  if ! go test -coverprofile $coverageFile.txt -failfast -v -p 1 $s;
    then break;
  fi;
done

go tool cover -html=$coverageFile.txt -o $coverageFile.html

# go test -v -p 1 -failfast \
#     $testFiles \
#     -coverpkg=github.com/fastschema/fastschema/... \
#     -coverprofile $scriptDir/coverage/coverage.txt ./... && go tool cover \
#       -html=$scriptDir/coverage/coverage.txt \
#       -o $scriptDir/coverage/coverage.html

cd $workingDir