FROM golang:1.13 as builder

COPY . /workdir/
WORKDIR /workdir

RUN make build

FROM debian:buster

COPY --from=builder /workdir/bin/cm500_exporter /cm500_exporter

ENTRYPOINT ["/cm500_exporter"]