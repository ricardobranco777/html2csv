FROM	docker.io/library/golang AS builder

WORKDIR	/go/src/html2csv
COPY	. .

RUN	make

FROM	scratch
COPY	--from=builder /go/src/html2csv/html2csv /usr/local/bin/html2csv

ENTRYPOINT ["/usr/local/bin/html2csv"]
