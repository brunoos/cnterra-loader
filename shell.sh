#!/bin/bash

mkdir -p tmp

docker run -it --rm \
 --device /dev/ttyUSB0 \
 --network cnterra-net \
 -v /home/brunoos/UFG/workspace/cnterra-loader/tmp:/opt/cnterra-loader/tmp \
 -v /home/brunoos/UFG/workspace/cnterra-loader/cnterra-loader:/opt/cnterra-loader/cnterra-loader \
 -v /home/brunoos/UFG/workspace/cnterra-loader/loader.sh:/opt/cnterra-loader/loader.sh \
 -v /home/brunoos/UFG/workspace/cnterra-loader/tos-bsl.py:/opt/cnterra-loader/tos-bsl.py \
 cnterra-loader-dev:1.0 \
 /bin/bash
