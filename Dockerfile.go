FROM golang@sha256:4d459123f9eae461060316b85091615a35684fe8a6d975e14aabbb12e9ddb528 AS builder
ENV TINI_VERSION=0.19.0
ADD https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini

RUN addgroup -g 10001 \
             -S nonroot && \
    adduser -u 10000 \
            -G nonroot \
            -h /home/nonroot \
            -S nonroot

WORKDIR /home/nonroot

COPY ./go.* ./
RUN go mod graph | awk '$1 !~ /@/ { print $2 }' | xargs -r go get -v

USER nonroot:nonroot

COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 go build -ldflags="-w -s" \
                          -gcflags=all="-l -B -C" \
                          -o ./app \
                          -v ./main.go

FROM gcr.io/distroless/static-debian12@sha256:22cda3953e236576e5d113fc6a9290b27e179fbe5937efb1fdc16d81ddfb6120
USER nonroot:nonroot

WORKDIR /home/nonroot

COPY --from=builder /tini /tini
ENTRYPOINT ["/tini", "--"]

COPY --from=builder /home/nonroot/app .
CMD ["./app"]