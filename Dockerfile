FROM ubuntu:18.04

RUN apt-get update ; apt-get upgrade -y

RUN apt-get install -y build-essential python python-serial
RUN apt-get install -y tinyos-tools gcc-avr gcc-msp430

COPY cnterra-loader      /opt/cnterra-loader/
COPY start.sh            /opt/cnterra-loader/
COPY loader.sh           /opt/cnterra-loader/
COPY tos-bsl             /opt/cnterra-loader/
COPY tos-bsl-license.txt /opt/cnterra-loader/

RUN chmod a+x /opt/cnterra-loader/*.sh
RUN chmod a+x /opt/cnterra-loader/cnterra-loader
