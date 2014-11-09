
SRC = cq.go sq.go squ.go

build:: ${SRC}
	go build ${BF}

install:: ${SRC}
	go install ${IF}

vet:: ${SRC}
	go vet ${VF}

test:: ${SRC}
	go test ${TF}

clean::
	rm -f ${SRC}
	rm -f *~
	rm -f *.out
	go clean ${CF}

cq.go : cq.gox
	./mkq.bash fixq CQ 'interface{}' < cq.gox > cq.go
	chmod -w cq.go

sq.go : sq.gox
	./mkq.bash fixq SQ 'interface{}' < sq.gox > sq.go
	chmod -w sq.go

squ.go : squ.gox
	./mkq.bash fixq SQU 'interface{}' < squ.gox > squ.go
	chmod -w squ.go
