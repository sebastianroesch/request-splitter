# STEP 1 build executable binary
FROM golang:alpine as builder

# Install git + SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates

# Create appuser
RUN adduser -D -g '' appuser
COPY . $GOPATH/src/github.com/sebastianroesch/request-splitter/
WORKDIR $GOPATH/src/github.com/sebastianroesch/request-splitter/

#get dependancies
RUN go get -d -v

#build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/request-splitter


# STEP 2 build a small image
# start from scratch
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /go/bin/request-splitter /go/bin/request-splitter
COPY config/config.production.json /config/config.json
ENV ENV=production
USER appuser
ENTRYPOINT ["/go/bin/request-splitter"]