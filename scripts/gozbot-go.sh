#!/bin/time /home/zosgo/zopen/bin/bash
cd gozbot 2>/dev/null || (
    echo "Not Setup" >&2
    exit 1
)

cleanup() {
        echo "Cleaning Up =============="
        cd ..
        rm -rf $goBuildDir
}

repo=$1
branch=$2
echo "repo: ${repo}"
echo "branch: ${branch}"

goBuildDir="build-${branch}"
echo "goBuildDir: ${goBuildDir}"

git clone -b $branch "git@github.ibm.com:${repo}" $goBuildDir
cd $goBuildDir
tagfile -rq .

echo "Go Build ================="
timeout 7200 "./go-build-zos/native_build.sh"
if [ $? -gt 0 ]; then
        cleanup
        exit $?
fi
eval $(./go-build-zos/bin/goz-env)
if [ $? -gt 0 ]; then
        cleanup
        exit $?
fi

echo "Go Tests ================="
cd ./src && ./run.bash
cd ..
if [ $? -gt 0 ]; then
        cleanup
        exit $?
fi

cleanup
exit 0
