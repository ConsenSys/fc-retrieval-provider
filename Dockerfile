# Copyright (C) 2020 ConsenSys Software Inc
FROM golang:1.16-alpine as builder
RUN apk update && apk add --no-cache make gcc musl-dev linux-headers git
WORKDIR /go/src/github.com/ConsenSys/fc-retrieval-provider/
COPY . .

# Get all dependancies, but don't install.
RUN go get -d -v github.com/ConsenSys/fc-retrieval-provider/cmd/provider
# Do a full compile of app and dependancies, forcing static linking.
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /go/bin/provider ./cmd/provider

# Pull build provider into a second stage deploy alpine container
FROM alpine:latest
COPY --from=builder /go/bin/provider /provider

# Run the binary when the container starts.
WORKDIR /
CMD ["/provider"]
EXPOSE 9030
