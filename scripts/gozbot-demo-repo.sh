#!/bin/bash
cd gozbot 2>/dev/null || (
    echo "Not Setup" >&2
    exit 1
)

cleanup() {
        echo "Cleaning Up =============="
        cd ..
        rm -rf $buildDir
}

repo=$1
branch=$2
echo "repo: ${repo}"
echo "branch: ${branch}"

buildDir="build-${branch}"
echo "buildDir: ${buildDir}"

git clone -b $branch "git@github.ibm.com:${repo}" $buildDir
cd $buildDir

echo "Go Build ================="
go build
if [ $? -gt 0 ]; then
        cleanup
        exit 1
fi

./goz-workflow-demo
if [ $? -gt 0 ]; then
        cleanup
        exit 1
fi

echo "Go Test =================="
go test
if [ $? -gt 0 ]; then
        cleanup
        exit 1
fi

cleanup
