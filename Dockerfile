FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o terraview

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Graphviz version 10.0.1
RUN curl -SL https://graphviz.gitlab.io/pub/graphviz/stable/ubuntu/ubuntu-22.04/graphviz_10.0.1-1_amd64.deb -o graphviz.deb \
    && apt-get update \
    && apt-get install -y ./graphviz.deb \
    && rm graphviz.deb \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/terraview /usr/local/bin/terraview

CMD ["sh"]