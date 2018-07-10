#cgo的dnslookup时，调用glibc getaddrinfo会存在bug
#可以通过netgo tag调用go 原生的dnslookup
export GOPATH=/opt/build/third-party
export GOPATH=$GOPATH:/opt/build/adx
go build -tags netgo adx.go
