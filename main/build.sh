#Build linux windows and binaries
export GOOS=linux
go build -o monkey monkey.go
export GOOS=windows
go build -o monkey.exe monkey.go