#!/bin/bash

mkdir -p tmp

docker run -it --rm \
 --device /dev/ttyUSB0 \
 --network cnterra-net \
 -v /opt/cnterra-loader:/opt/cnterra-loader \
 cnterra-loader-dev:1.0 \
 /bin/bash
