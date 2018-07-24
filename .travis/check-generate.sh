#!/bin/bash

# Add any go generate invocations below this line.


if [ -n "$(git diff)" ]; then
	git diff
	exit 1
fi
