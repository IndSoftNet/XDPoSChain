.PHONY: XDC XDC-cross evm all test clean
.PHONY: XDC-linux XDC-linux-386 XDC-linux-amd64 XDC-linux-mips64 XDC-linux-mips64le
.PHONY: XDC-darwin XDC-darwin-386 XDC-darwin-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

XDC:
	$(GORUN) build/ci.go install ./cmd/XDC
	@echo "Done building."
	@echo "Run \"$(GOBIN)/XDC\" to launch XDC."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/XDC.aar\" to use the library."
	@echo "Import \"$(GOBIN)/XDC-sources.jar\" to add javadocs"
	@echo "For more info see https://stackoverflow.com/questions/20994336/android-studio-how-to-attach-javadoc"
	
ios:
	$(GORUN) build/ci.go xcode --local

gc:
	build/env.sh go run build/ci.go install ./cmd/gc
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gc\" to launch gc."

bootnode:
	build/env.sh go run build/ci.go install ./cmd/bootnode
	@echo "Done building."
	@echo "Run \"$(GOBIN)/bootnode\" to launch a bootnode."

puppeth:
	build/env.sh go run build/ci.go install ./cmd/puppeth
	@echo "Done building."
	@echo "Run \"$(GOBIN)/puppeth\" to launch puppeth."

all:
	build/env.sh go run build/ci.go install

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# Cross Compilation Targets (xgo)

XDC-cross: XDC-linux XDC-darwin
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/XDC-*

XDC-linux: XDC-linux-386 XDC-linux-amd64 XDC-linux-mips64 XDC-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-*

XDC-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/XDC
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep 386

XDC-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/XDC
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep amd64

XDC-linux-arm: XDC-linux-arm-5 XDC-linux-arm-6 XDC-linux-arm-7 XDC-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep arm

XDC-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/XDC
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep arm-5

XDC-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/XDC
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep arm-6

XDC-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/XDC
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep arm-7

XDC-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/XDC
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep arm64

XDC-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/XDC
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep mips

XDC-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/XDC
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep mipsle

XDC-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/XDC
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep mips64

XDC-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/XDC
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/XDC-linux-* | grep mips64le

XDC-darwin: XDC-darwin-386 XDC-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/XDC-darwin-*

XDC-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/XDC
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-darwin-* | grep 386

XDC-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/XDC
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-darwin-* | grep amd64

XDC-windows: XDC-windows-386 XDC-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/XDC-windows-*

XDC-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/XDC
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-windows-* | grep 386

XDC-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/XDC
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/XDC-windows-* | grep amd64
