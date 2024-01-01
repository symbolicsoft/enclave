windows:
	@mkdir build/windows
	@GOOS="windows" GOARCH="amd64" go build -trimpath -gcflags="-e" -ldflags="-s -w" -o -gcflags=all="-l -B" -o build/windows/enclave.exe github.com/symbolicsoft/enclave/v2/cmd/enclave
	@cd build/windows; upx --best enclave

linux:
	@mkdir build/linux
	@GOOS="linux" GOARCH="amd64" go build -trimpath -gcflags="-e" -ldflags="-s -w" -gcflags=all="-l -B" -o build/linux/enclave github.com/symbolicsoft/enclave/v2/cmd/enclave
	@cd build/linux; upx --best enclave

macos:
	@mkdir build/macos
	@GOOS="darwin" GOARCH="arm64" go build -trimpath -gcflags="-e" -ldflags="-s -w" -o -gcflags=all="-l -B" -o build/macos/enclave github.com/symbolicsoft/enclave/v2/cmd/enclave
	@cd build/macos; upx --best enclave

protobuf:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@cd internal; protoc --go_out=. --go-grpc_out=. enclave.proto

update:
	@$(RM) go.sum
	@go get -u all
	@go mod tidy

clean:
	@$(RM) -rf build/windows build/linux build/macos

.PHONY: windows linux macos protobuf update build cmd internal
