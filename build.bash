#!/bin/bash -eu
#
# Compile gocryptfs and bake the git version string of itself and the go-fuse
# library into the binary.
#
# If you want to fake a build date to reproduce a specific build,
# you can use:
# BUILDDATE=2017-02-03 ./build.bash

cd "$(dirname "$0")"
MYDIR=$PWD

# Make sure we have the go binary
go version > /dev/null

# GOPATH may contain multiple paths separated by ":"
GOPATH1=$(go env GOPATH | cut -f1 -d:)

if [[ $PWD != *"/src/github.com/rfjakob/gocryptfs" ]] ; then
	echo "Warning: Building outside of GOPATH will most likely fail."
	echo "         Please rename $PWD to $GOPATH1/src/github.com/rfjakob/gocryptfs ."
	sleep 5
	echo
fi

# gocryptfs version according to git or a VERSION file
if [[ -d .git ]] ; then
	GITVERSION=$(git describe --tags --dirty)
elif [[ -f VERSION ]] ; then
	GITVERSION=$(cat VERSION)
else
	echo "Warning: could not determine gocryptfs version"
	GITVERSION="[unknown]"
fi

# go-fuse version, if available
if [[ -d vendor/github.com/hanwen/go-fuse ]] ; then
	GITVERSIONFUSE="[vendored]"
else
	# go-fuse version according to git
	# Note: git in CentOS 7 does not have "git -C" yet, so we use plain "cd".
	FAIL=0
	cd $GOPATH1/src/github.com/hanwen/go-fuse
	OUT=$(git describe --tags --dirty 2>&1) || FAIL=1
	if [[ $FAIL -eq 0 ]]; then
		GITVERSIONFUSE=$OUT
	else
		echo "$PWD: git describe: $OUT"
		echo "Warning: could not determine go-fuse version"
		GITVERSIONFUSE="[unknown]"
	fi
	cd "$MYDIR"
fi

# Build date, something like "2017-09-06"
if [[ -z ${BUILDDATE:-} ]] ; then
	BUILDDATE=$(date +%Y-%m-%d)
fi

LDFLAGS="-X main.GitVersion=$GITVERSION -X main.GitVersionFuse=$GITVERSIONFUSE -X main.BuildDate=$BUILDDATE"
go build "-ldflags=$LDFLAGS" $@

(cd gocryptfs-xray; go build $@)

./gocryptfs -version

mkdir -p $GOPATH1/bin
cp -af gocryptfs $GOPATH1/bin
