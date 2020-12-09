#!/bin/sh

if [ "${1:0:1}" = '-' ]; then
    set -- ensemble "$@"
fi

# Look for ensemble subcommands.
if [ "$1" = 'server' ]; then
    shift
    set -- ensemble server "$@"
fi

exec "$@"