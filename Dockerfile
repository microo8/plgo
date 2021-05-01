FROM ubuntu:20.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    wget \
    curl \
    ca-certificates \
    g++ \
    gcc \
    libc6-dev \
    make \
    pkg-config \
    gnupg

# Install PostgreSQL development tools
RUN set -e && \
    echo "deb http://apt.postgresql.org/pub/repos/apt focal-pgdg main" >> /etc/apt/sources.list.d/pgdg.list && \
    curl https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update && \
    apt-get -y install postgresql-client-13 postgresql-server-dev-13 && \
    rm -rf /var/lib/apt/lists/*

# Install Go
RUN set -e && \
    wget -q -O go.tgz "https://golang.org/dl/go1.16.3.linux-amd64.tar.gz"; \
    tar -C /usr/local -xzf go.tgz; \
    rm go.tgz; \
    export PATH="/usr/local/go/bin:$PATH";

ENV GOPATH /usr/local/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install PLGO, using "go get" would actually place this in the pkg folder since
# it is a module, but the writer expects the PLGO source to be located in the
# src folder for now.
COPY . /usr/local/go/src/github.com/paulhatch/plgo

RUN cd /usr/local/go/src/github.com/paulhatch/plgo && \
    go get -d ./... && \
    go install github.com/paulhatch/plgo/plgo

