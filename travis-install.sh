#!/bin/bash
# travis does a build for for multiple versions of go
# at this time the build from tip uses a new location for
# the cover tool, this script takes care of that

go get -t github.com/FreekKalter/km
go get -t github.com/FreekKalter/km/lib

go_version=$(go version)
if [[ $(echo "$go_version" | grep -c "devel") -gt 0 ]]; then
    go get -v golang.org/x/tools/cmd/cover
elif [[ $(echo "$go_version" | grep -c "go1.4") -gt 0 ]]; then
    go get -v golang.org/x/tools/cmd/cover
else
    go get -v code.google.com/p/go.tools/cmd/cover
fi
