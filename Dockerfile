FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o terraview

FROM ubuntu:24.10

RUN apt-get -y update

# Install Graphviz version 10.0.1
RUN apt-get install -y graphviz

# Install Terraform 1.8.5
RUN curl -LO https://releases.hashicorp.com/terraform/1.8.5/terraform_1.8.5_linux_amd64.zip \
    && apt-get install -y unzip \
    && unzip terraform_1.8.5_linux_amd64.zip \
    && mv terraform /usr/local/bin/ \
    && rm terraform_1.8.5_linux_amd64.zip

COPY --from=builder /app/terraview /usr/local/bin/terraview

CMD ["sh"]