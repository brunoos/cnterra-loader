FROM ubuntu:18.04

RUN apt-get update ; apt-get upgrade -y

RUN apt-get install -y build-essential python python-serial
RUN apt-get install -y tinyos-tools gcc-avr gcc-msp430
RUN apt-get install -y net-tools uuid-runtime vim