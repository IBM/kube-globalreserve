FROM debian:stretch-slim

WORKDIR /

COPY kube-globalreserve-scheduler /usr/local/bin

CMD ["kube-globalreserve-scheduler"]
