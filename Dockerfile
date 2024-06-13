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
RUN apt install graphviz=10.0.1

COPY --from=builder /app/terraview /usr/local/bin/terraview

CMD ["sh"]