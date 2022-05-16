#!/bin/bash

ID=1
PORT=8080
SERIAL=/dev/ttyUSB0
RABBITMQ=cnterra-rabbitmq
DIR=`pwd`

docker run -it --rm \
 --device ${SERIAL} \
 --network cnterra-net \
 -v ${DIR}:/opt/cnterra-loader \
 -p ${PORT}:${PORT} \
 -e NODE_ID=${ID} \
 -e RABBITMQ_ADDRESS=${RABBITMQ} \
 -e LOADER_PORT=${PORT} \
 -e SERIAL_PORT=${SERIAL} \
 cnterra-loader-dev:1.0 \
 /bin/bash
