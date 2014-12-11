#!/bin/bash
# travis does a build for for multiple versions of go
# at this time the build from tip uses a new location for
# the cover tool, this script takes care of that

go get -t github.com/FreekKalter/km
go get -t github.com/FreekKalter/km/lib

go version | grep devel &>/dev/null
go_version=$(go version)
if [[ $(echo "$version" | grep -qc "devel") -ge 0 ]]; then
    go get -v golang.org/x/tools/cmd/cover
elif [[ $(echo "$version" | grep -qc "go1.4") -ge 0 ]]; then
    go get -v golang.org/x/tools/cmd/cover
else
    go get -v code.google.com/p/go.tools/cmd/cover
fi
