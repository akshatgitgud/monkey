#Build linux windows and binaries
export GOOS=linux
go build -o monkey .

# Dont need the windows one for now
# export GOOS=windows
# go build -o monkey.exe monkey.go