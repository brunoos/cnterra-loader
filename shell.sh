#!/bin/bash

DIR=`pwd`

docker run -it --rm \
 --device /dev/ttyUSB0 \
 --network cnterra-net \
 -v ${DIR}:/opt/cnterra-loader \
 -p 8080:8080 \
 -p 8081:8081 \
 cnterra-loader-dev:1.0 \
 /bin/bash
