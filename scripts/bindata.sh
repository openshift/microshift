go get -u github.com/go-bindata/go-bindata/...
for i in crd core scc components; do
	OUTPUT="pkg/assets/${i}/bindata.go"
	${GOPATH}/bin/go-bindata -nocompress -prefix "pkg/assets/${i}" -pkg assets -o ${OUTPUT} "./assets/${i}/..."
	gofmt -s -w "${OUTPUT}"
done
