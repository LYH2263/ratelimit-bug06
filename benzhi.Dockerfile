FROM golang:1.22
WORKDIR /src
COPY . .
RUN go test ./... -count=1
