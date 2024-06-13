FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o terraview

FROM ubuntu:24.10

RUN apt-get -y update

# Install Graphviz version 10.0.1
RUN apt-get install -y graphviz

COPY --from=builder /app/terraview /usr/local/bin/terraview

CMD ["sh"]