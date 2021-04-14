FROM ubuntu:20.04

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates git \
    zopfli jpegoptim curl
RUN update-ca-certificates

ARG NODE_VERSION=14.16.0
ARG NODE_PACKAGE=node-v$NODE_VERSION-linux-x64
ARG NODE_HOME=/opt/$NODE_PACKAGE

ENV NODE_PATH $NODE_HOME/lib/node_modules
ENV PATH $NODE_HOME/bin:$PATH

RUN curl https://nodejs.org/dist/v$NODE_VERSION/$NODE_PACKAGE.tar.gz | tar -xzC /opt/

ENV BOT_BASE_PATH=/cdnjs

RUN mkdir -p /cdnjs \
             /cdnjs/cdnjs \
             /cdnjs/glob \
             /cdnjs/packages

RUN cd /cdnjs/cdnjs && \
    git init

COPY dev/packages /cdnjs/packages/packages

RUN git clone https://github.com/cdnjs/glob.git /cdnjs/glob
RUN cd /cdnjs/glob && npm install

COPY . /cdnjs/tools
COPY bin/autoupdate /usr/bin/autoupdate
RUN cd /cdnjs/tools && npm install

ENV WORKERS_KV_ACCOUNT_ID=empty \
    WORKERS_KV_AGGREGATED_METADATA_NAMESPACE_ID=empty \
    WORKERS_KV_API_TOKEN=empty \
    WORKERS_KV_FILES_NAMESPACE_ID=empty \
    WORKERS_KV_PACKAGES_NAMESPACE_ID=empty \
    WORKERS_KV_SRIS_NAMESPACE_ID=empty \
    WORKERS_KV_VERSIONS_NAMESPACE_ID=empty \
    ALGOLIA_WRITE_API_KEY=empty \
    DEBUG=1

RUN git config --global user.email "dev@cdnjs.com"
RUN git config --global user.name "dev"

WORKDIR /cdnjs/cdnjs
