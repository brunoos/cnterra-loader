#!/bin/bash

# Usage: loader.sh <serial> <file.exe>

PORT="$1"
FILE="$2"
IHEX="${FILE}.ihex"
IHEXID="id-${IHEX}"

msp430-objcopy --output-target=ihex $FILE $IHEX

#tos-set-symbols --objcopy msp430-objcopy --objdump msp430-objdump --target ihex $IHEX $IHEXID TOS_NODE_ID=$2 ActiveMessageAddressC__addr=$2 

python2.7 /opt/cnterra-loader/tos-bsl --telosb -c $PORT -r -e -I -p $IHEX
