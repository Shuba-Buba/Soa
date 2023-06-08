FROM golang:latest

WORKDIR /hw1/data
COPY . /hw1/data

ENV GROUP_ADDRESS 239.0.0.1:54321

ENTRYPOINT ["go", "run", "main.go"]