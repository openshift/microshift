go install github.com/go-bindata/go-bindata/...
for i in crd core rbac apps scc storage; do
	OUTPUT="pkg/assets/${i}/bindata.go"
	"${GOPATH}"/bin/go-bindata -nocompress -nometadata -prefix "pkg/assets/${i}" -pkg assets -o ${OUTPUT} "./assets/${i}/..."
	gofmt -s -w "${OUTPUT}"
done
