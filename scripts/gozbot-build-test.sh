#!/bin/time /home/zosgo/zopen/bin/bash
if [[ -z $2 ]] && ! [[ -z $1 ]]; then
        echo "Likely invoked by travis, return to bypass"
        exit 0
fi

if ! [[ -x ~/parseGoTest ]]; then
        echo "No parseGoTest found. Stopping"
        exit 1
fi

trap "exit 0" USR1
trap "exit -1" USR2
echo "pid:$$ arguments: $@"
export TERM=dtterm
export _BPXK_AUTOCVT=ALL
export _CEE_RUNOPTS="FILETAG(AUTOCVT,AUTOTAG),HEAPPOOLS(ON),TERMTHDACT(UADUMP),ABTERMENC(ABEND),HEAPPOOLS64(ON,(256,4),20,(512,2),10)"
export _TAG_REDIR_ERR=txt
export _TAG_REDIR_IN=txt
export _TAG_REDIR_OUT=txt
unset dir
if [[ -x ~/zopen/usr/local/bin/location_dir ]]; then
        dir=$(~/zopen/usr/local/bin/location_dir)
else
        if [[ -x /home/opnzos/zopen/usr/local/bin/location_dir ]]; then
                dir=$(/home/opnzos/zopen/usr/local/bin/location_dir)
        fi
fi
if ! [[ -z "${dir}" ]]; then
        ROOT="${dir}/../../../location_dir"
        export ZOPEN_ROOTFS=$(${dir}/../../../location_dir)
        if [[ -r "${ZOPEN_ROOTFS}/etc/zopen-config" ]]; then
                export ZOPEN_QUICK_LOAD=1
                . "${ZOPEN_ROOTFS}/etc/zopen-config"
                unset ZOPEN_QUICK_LOAD=
                export ASCII_TERMINFO="${ZOPEN_ROOTFS}/usr/local/share/terminfo"
                export TERM=xterm-256color
                export GIT_SSL_CAINFO="${ZOPEN_ROOTFS}/etc/pki/tls/certs/cacert.pem"
                export SSL_CERT_FILE=$GIT_SSL_CAINFO
        fi
fi
export TIMEOUTCMD="$HOME/zopen/bin/timeout 7200"
export CLEANUP="$HOME/zopen/bin/cleanup"

tagfile=$HOME/zopen/bin/tagfile
if ! [[ -x "$CLEANUP" ]]; then
        export CLEANUP="echo"
        echo "WARNING cleanup not found"
        echo "WARNING cleanup not found" >&2
else
        $CLEANUP /home/zosgo/go-build /home/zosgo/build1 /home/zosgo/build2 /home/zosgo/build3 /home/zosgo/tmp /home/zosgo/.cache /tmp
fi

if [[ -r $HOME/.ssh.env ]]; then
        source $HOME/.ssh.env
fi

id=${LOGNAME}
export _BPXK_MDUMP="${id}.DUMP"

source "${HOME}/.bashrc"

ispid() {
        local re='^[0-9]+$'
        if [[ "${1}" =~ $re ]]; then
                kill -0 "${1}" 2>/dev/null
                if [[ $? -eq 0 ]]; then
                        true
                else
                        false
                fi

        else
                false
        fi
}
build_dir="${HOME}/go-build"
mkdir -p $build_dir

if [[ -d $HOME/build1 ]] && [[ -d $HOME/build2 ]] && [[ -d $HOME/build3 ]]; then
        #on zoscan56 we have workspace pool
        export NOLOCK=1
        declare -A pool
        declare -A dfpool
        $CLEANUP $HOME/build1 $HOME/build2 $HOME/build3 $HOME/go-build

        pool["$HOME/build1"]="$HOME/build1"
        pool["$HOME/build2"]="$HOME/build2"
        pool["$HOME/build3"]="$HOME/build3"
        pool["$HOME/go-build"]="$HOME"

        while IFS=" " read -r -a line; do
                dfpool[${line[5]}]=${line[3]}
        done < <(/bin/df -P -k 2>/dev/null | /bin/tail -n +2)

        max=0
        res=""
        for key in ${!pool[@]}; do
                key1=${pool[$key]}
                echo "directory $key1 has ${dfpool[$key1]}k free"
                if [[ ${dfpool[$key1]} -gt $max ]]; then
                        max=${dfpool[$key1]}
                        res=$key
                fi
        done
        build_dir="${res}"
        echo build directory pool selected ${build_dir}
fi
if [[ -z "$NOLOCK" ]]; then
        LOCKDIR=$HOME/lockdir
        /bin/mkdir "${LOCKDIR}" 2>/dev/null
        while [[ $? -ne 0 ]]; do
                opid=$(cat "${LOCKDIR}/pid" 2>/dev/null)
                ispid "$opid"
                if [[ $? -eq 0 ]]; then
                        echo "Waiting for $opid to finish"
                        sleep 30
                else
                        /bin/rm -rf "${LOCKDIR}"
                fi
                /bin/mkdir "${LOCKDIR}" 2>/dev/null
        done
fi
trap - USR1
trap - USR2
if [[ -z "$NOLOCK" ]]; then
        trap "rm -f ${LOCKDIR}/pid; rmdir ${LOCKDIR}; exit 0" USR1
        trap "rm -f ${LOCKDIR}/pid; rmdir ${LOCKDIR}; exit -1" USR2
        trap "rm -f ${LOCKDIR}/pid; rmdir ${LOCKDIR}; exit" INT TERM EXIT
        echo $$ >${LOCKDIR}/pid
else
        trap "exit 0" USR1
        trap "exit -1" USR2
        trap "exit" INT TERM EXIT
fi
declare -A cleanlist

tsocmd "delete '${_BPXK_MDUMP}'"
tsocmd "allocate da('${_BPXK_MDUMP}') space(900,50) cyl RECFM(F,B,S) lrecl(4160) blksize(4160) storclas(DUMPS) new"

me=$0
rc=1
if [[ ${me:0:1} != '/' ]]; then
        me="$(pwd)/$me"
fi
killtree=$(type -p killtree.sh)

cleanup() {
        cd /tmp
        for f in ${!cleanlist[@]}; do
                if [[ -r ${f} ]]; then
                        echo cleanup ${f}
                        rm -rf ${f}
                fi
        done
}

bootstrap=${GOROOT_BOOTSTRAP:=/home/zosgo/bootstrap}

die() {
        echo $@
        exit 1
}
newdir() {
        local ndir
        local prefix="$1"
        for i in {1..1000}; do
                if ! [[ -e "${prefix}-$i" ]]; then
                        ndir="${prefix}-$i"
                        mkdir -p "${ndir}" || die "cannot create directory ${ndir}"
                        cleanlist["${ndir}"]=1
                        echo "${ndir}"
                        return
                fi
        done
        die "cannot create new directory with prefix $prefix"
        false
}
function timeout() {
        timeout=$1
        wakeup=1
        loopcount=7200
        count=$loopcount
        child_dead=0
        shift
        (
                eval $@ &
                child=$!
                trap -- "" SIGTERM
                (
                        exec 0</dev/null
                        exec 1>/dev/null
                        while [[ $timeout -gt 0 ]]; do
                                kill -0 $child 2>/dev/null
                                if [[ $? -ne 0 ]]; then
                                        child_dead=1
                                        break
                                fi
                                ((count -= 1))
                                if [[ $timeout -ge $wakeup ]]; then
                                        ((timeout -= $wakeup))
                                        sleep $wakeup
                                        if [[ count -eq 0 ]]; then
                                                #pacify travis
                                                echo -ne "." >&2
                                                count=$loopcount
                                        fi
                                else
                                        sleep $timeout
                                        timeout=0
                                fi
                        done
                        if [[ $child_dead -eq 0 ]]; then
                                $killtree $child 2>/dev/null
                        fi
                ) &
                wait $child
        )
}

mkdir -p $build_dir

cd $build_dir 2>/dev/null || die "failed to cd to $build_dir"

repo=$(newdir "$build_dir/build")
GOCACHE=$(newdir "$build_dir/gocache")
export GOCACHE
TMPDIR=$(newdir "$build_dir/tmpdir")
export TMPDIR
export GOTOOLCHAIN=local

! [[ -z "$repo" ]] || die "no suitable repository directory"

if ! [[ -z "$1" ]]; then
        BR="-b $1"
else
        unset BR
fi

git clone --recurse-submodules git@github.ibm.com:open-z/go.git $BR $repo || die "git clone --recurse-submodules git@github.ibm.com:open-z/go.git failed"

cd $repo || die "cd $repo failed"

git config user.name "build"
git config user.email "$(id -nu)@$(hostname)"

if [[ -z "$1" ]]; then
        echo -ne "\nbranch not provided, use default\n\n"
else
        if [[ -z "$2" ]]; then
                git checkout "$1" --recurse-submodule -f || die "git checkout $1 failed"
        else
                if [[ "$1" == "$2" ]]; then
                        git checkout "$1" --recurse-submodule -f || die "git checkout $1 failed"
                else
                        git checkout "$2"
                        git branch -D "$1"
                        git fetch origin || die "git fetch origin failed"
                        git checkout --recurse-submodules -f -b "$1" "origin/$1" || die "git checkout -f -b $1 origin/$1 failed"
                        git merge "origin/$2" -m "Merge-test" || die "git merge $2 failed"
                fi
        fi
fi

branch=$(git status -b --porcelain | head -1 | awk -F / '{print $2}') || die "unknown branch"
#
#
if false; then
        if [[ -r ./VERSION ]]; then
                buildver=$(/bin/cat ./VERSION)
                if [[ ${buildver:0:6} == "go1.18" ]]; then
                        SKIPTEST=1
                fi
        fi
fi
#
#

version=$(git rev-parse --short HEAD 2>/dev/null)
git update-index --refresh -q >/dev/null
dirty=$(git diff-index --name-only HEAD 2>/dev/null)
if [ -n "$dirty" ]; then
        version="$version-dirty"
fi
paxfile="${branch}-$(date +%Y-%m-%d)-${version}.pax.Z"
paxpath="$(pwd)/../${paxfile}"

lscommits() {
        u=$(grep \"username\" $HOME/.artifactory.json)
        p=$(grep \"sectoken\" $HOME/.artifactory.json)
        u=${u#*\"username\":*\"}
        p=${p#*\"sectoken\":*\"}
        u=${u%%\"*}
        p=${p%%\"*}
        local afurl=$(dirname $(go-artifact -i "${1}" -v -q 2>&1 | grep -E "^HEAD " | awk '{print $3}'))
        curl -s --user $u:$p -k "${afurl}/" | grep "a href=\"[^.]" | /bin/sed -e "s/^.*<a href=\".*-\([0-9a-fA-F]*\)\.pax.*\".*a>\(.*\)/\1/g"
}

cm=$(lscommits "${paxfile}")
for c in $cm; do
        if [[ ${version} == $c ]]; then
                echo "This version ${version} is on artifactory, no need to rebuild"
                exit 0
        fi
done

# always upload
do_upload=1

echo -ne "\nStart build of $paxpath\n\n"

t0=$(awk 'BEGIN {srand(); print srand()}')

export COMPILER_PATH=/home/opnzos/local/bin
export CGO_ENABLED=1

# retry the build a couple of times
retries=2
if [[ "$(type -p storeenv)" != "" ]]; then
        storeenv >$HOME/.buildenv
fi
$tagfile -rq .
VER=$(cat ./VERSION)
if [[ $VER =~ ^go1.1[6-8] ]]; then
        export COMPILER_PATH=/home/opnzos/local/xlcpp/bin
        bootstrap=/home/opnzos/local/go1.18.4
fi
if [[ $VER =~ ^go1.19 ]]; then
        export COMPILER_PATH=/home/opnzos/local/bin
        bootstrap=/home/opnzos/local/1.19.4
fi

# Determine the proper build script (for backward compat reasons)
zos_util_dir=misc/zos
build_script_path=misc/zos/build.sh
pax_script_path=misc/zos/scripts/pax.sh
sanity_script_path=misc/zos/test/quicksanity.mak
if [[ ! -f "misc/zos/build.sh" ]]; then
        zos_util_dir=go-build-zos
        build_script_path=go-build-zos/native_build.sh
        pax_script_path=go-build-zos/native_create_pax.sh
        sanity_script_path=go-build-zos/quicksanity.mak
fi

#temp bypass so we can run  the right set of tests.
if ! [[ -d go-build-zos ]]; then
        git clone git@github.ibm.com:open-z/go-build-zos.git
fi
zos_util_dir=go-build-zos
build_script_path=go-build-zos/native_build.sh
pax_script_path=go-build-zos/native_create_pax.sh
sanity_script_path=go-build-zos/quicksanity.mak

# unboring begins
currdir=$(pwd)
while read line; do
        cd "$line" || exit 1
done < <(git rev-parse --show-toplevel)
declare -a list
list=(
        src/crypto/internal/boring/Dockerfile
        src/crypto/internal/boring/LICENSE
        src/crypto/internal/boring/build.sh
        src/crypto/internal/boring/goboringcrypto.h
        src/crypto/internal/boring/syso/goboringcrypto_linux_amd64.syso)

for f in ${list[@]}; do
        dn=$(/bin/dirname $f)
        if ! [[ -d "${dn}" ]]; then
                echo $dn is not a directory, something is not right
        fi
        rm -f "${f}"
done
# unboring ends

# Run build
export GOMAXPROCS=1
export LIBPATH=/usr/lib:$LIBPATH
timeout 7200 "${build_script_path}" --goroot_bootstrap="${bootstrap}"
while [[ $? -ne 0 ]] && [[ $retries -gt 0 ]]; do
        echo "Retrying build"
        ((retries -= 1))
        timeout 7200 "${build_script_path}" --goroot_bootstrap="${bootstrap}"
done

if [[ $? -ne 0 ]]; then
        die "build failed"
fi
unset GOMAXPROCS
git status
git checkout --recurse-submodule -f
git clean -fd

t1=$(awk 'BEGIN {srand(); print srand()}')
((diff = t1 - t0))
echo "Elapse time $diff seconds"

# Compress binaries
for f in bin/* ./$zos_util_dir/bin/clang-wrapper* ./$zos_util_dir/bin/goz-env ./pkg/tool/zos_s390x/*; do
        echo compressing "$f"
        $zos_util_dir/bin/goz-util -c "$f"
done

# Create the PAX archive
echo -ne "\nCreate pax.Z archive\n\n"
$pax_script_path --paxpath="${paxpath}" || die "fail to create ${paxpath}"

cd ..

GOCACHE=$(newdir "$build_dir/gocache")
export GOCACHE
TMPDIR=$(newdir "$build_dir/tmpdir")
export TMPDIR
testdir=$(newdir "$build_dir/test")

cd "$testdir" 2>/dev/null || die "cd to $testdir failed"

pax -ppx -rf "$paxpath" || die "pax unpack failed"

export LIBPATH=/usr/lib:$LIBPATH

echo -ne "\nSanity test archive\n\n"

mkdir -p go/go-build-zos/
cp "${repo}/${sanity_script_path}" go/go-build-zos/quicksanity.mak || die "failed to copy quicksanity.mak"

"$CLEANUP" "/tmp" "$build_dir"
mkdir -p runtest
cd runtest || die "cd failed"

type -p make
if [[ -x ../go/etc/goz-env ]]; then
        eval $(../go/etc/goz-env)
else
        . ../go/etc/envsetup
fi
/home/opnzos/local/bin/storeenv
go clean
go clean -cache
go clean -modcache

export PATH="${build_dir}/${repo}/${zos_util_dir}/bin:${PATH}"
# retry the fail ones a couple of times
retries=5
if [[ "$SKIPTEST" == "1" ]]; then
        rc=0
        true
else
        echo "Starting Quicksanity. This might take a while..."
        storeenv >make.env
        timeout 7200 make -k -j 20 -f "../go/go-build-zos/quicksanity.mak" > /dev/null 2>&1
        while [[ $? -ne 0 ]] && [[ $retries -gt 0 ]]; do
                echo "Retrying Tests..."
                ((retries -= 1))
                storeenv >make.env
                echo "cd $(pwd)" >>make.env
                timeout 7200 make -k -j 20 -f "../go/go-build-zos/quicksanity.mak" > /dev/null 2>&1
                #-------------------------
                RC=$?
                while [[ $INSPECT ]] && [[ "$(type -p storeenv)" != "" ]] && [[ $RC -ne 0 ]]; do
                        echo make -k -f "../go/go-build-zos/quicksanity.mak" >>make.env
                        echo Wait for 3600 second
                        echo in $(pwd)/make.env
                        sleep 3600
                        [[ -r .env ]] && . ./.env
                done
                if [[ $RC -eq 0 ]]; then
                        true
                else
                        false
                fi
                #-------------------------

        done

fi

# Test output formatting
~/parseGoTest .

rc=$?
fail=0
for f in $(ls *.tmp 2>/dev/null); do
        [[ $fail -eq 0 ]] && fail=1
#       echo $f
#       echo "=========================="
#       /bin/grep FAIL $f
#       if [[ $? -ne 0 ]]; then
#               /bin/cat $f
#       fi
done
if [[ $fail -eq 0 ]] && [[ $rc -eq 0 ]] && [[ $do_upload -eq 1 ]]; then
        chmod go+r "${paxpath}"
        chmod go+r "${paxpath}.ilan"
        /home/opnzos/local/bin/go-artifact -i "${paxpath}" -v -c 2>&1 | cat
        /home/opnzos/local/bin/go-artifact -i "${paxpath}.ilan" -v -c 2>&1 | cat
        rc=0
fi
disown $(jobs -p) 2>/dev/null
#echo "Number of failures: $fail"
echo "Test directory $(pwd) on $(hostname)"

if [[ $fail -eq 0 ]] && [[ $rc -eq 0 ]]; then
        cleanup
        exit 0
fi
exit 1
