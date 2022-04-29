#!/bin/bash

mountDir="/host"
path="/usr/sbin:/usr/bin:/sbin:/bin"
cmd=`basename "$0"`

if [ ! -d "${mountDir}" ]; then
    echo "The directory ${mountDir} does not exist in container"
    exit 1
fi
exec chroot ${mountDir} /usr/bin/env -i PATH="${path}" ${cmd} "${@:1}"
