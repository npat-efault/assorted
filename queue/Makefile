
SRC = cq.go sq.go squ.go

build:: ${SRC}
	go build ${GFLAGS}

install:: ${SRC}
	go install ${GFLAGS}

vet:: ${SRC}
	go vet ${GFLAGS}

test:: ${SRC}
	go test ${GFLAGS}

clean::
	rm -f ${SRC}
	rm -f *~
	rm -f *.out
	go clean ${GFLAGS}

cq.go : cq.gox
	./mkq.bash queue CQ 'interface{}' < cq.gox > cq.go
	chmod -w cq.go

sq.go : sq.gox
	./mkq.bash queue SQ 'interface{}' < sq.gox > sq.go
	chmod -w sq.go

squ.go : squ.gox
	./mkq.bash queue SQU 'interface{}' < squ.gox > squ.go
	chmod -w squ.go
