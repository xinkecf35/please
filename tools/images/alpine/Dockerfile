FROM alpine:3.11
MAINTAINER peter.ebden@gmail.com

# Most dependencies; Python, Java, Clang
RUN apk update && \
    apk add python3 py3-pip openjdk11 curl unzip git pkgconf go bash

# Use bash as the shell
ENTRYPOINT /bin/bash

# Custom plzconfig
COPY /plzconfig /etc/plzconfig

# Welcome message
COPY /motd.txt /etc/motd
RUN echo 'cat /etc/motd' >> /root/.bashrc
WORKDIR /tmp
