FROM debian:stretch-slim

WORKDIR /

COPY bin/kube-globalreserve-scheduler /usr/local/bin

CMD ["kube-globalreserve-scheduler"]
