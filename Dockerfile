FROM golang as builder
RUN apt-get update && apt-get install -y upx
ADD . /iso
RUN cd /iso && make build

FROM scratch
ENV BHOJPUR_ISO_NOLOCK=true
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /iso/iso /usr/bin//bhojpur/isomgr

ENTRYPOINT ["/usr/bin/bhojpur/isomgr"]