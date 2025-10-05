#!/usr/bin/env bash

if command -v node >/dev/null 2>&1; then
	echo "node executable found at $(command -v node)"
else
	echo "node executable not found"
fi
if [[ -x $(node) && -f $(node) ]]
then
    echo "File '$file' is executable"
else
    echo "File '$file' is not executable or found"
fi
