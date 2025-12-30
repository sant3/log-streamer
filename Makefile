# Final package name and destination folder
PACKAGE_NAME=file-streamer
DIST_DIR=dist
ARCH64=amd64

# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RM := powershell -Command Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
    DATE := $(shell powershell -Command "Get-Date -Format 'yyyy/MM/dd'")
    MKDIR := powershell -Command New-Item -ItemType Directory -Force -Path
    TAR := tar
    ZIP_CMD := powershell -Command Compress-Archive -Force
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        DETECTED_OS := Linux
    endif
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS := macOS
    endif
    RM := rm -rf
    DATE := $(shell date +%Y/%m/%d)
    MKDIR := mkdir -p
    TAR := tar
    ZIP_CMD := zip
endif

.PHONY: hello dist-fe dist-be dist-be-linux dist-be-darwin dist-be-windows generate-self-signed-cert

hello:
	@echo "Starting build on $(DETECTED_OS)..."

dist-fe:
	@echo "Run pnpm build"
	cd frontend && $(RM) node_modules && pnpm install && pnpm run build -- --output-path=./dist/streamer
ifeq ($(OS),Windows_NT)
	cd frontend/dist && $(TAR) -czf streamer.tar.gz streamer && $(RM) streamer
else
	cd frontend/dist && tar czf streamer.tar.gz streamer && rm -rf streamer
endif
	@echo "Created streamer.tar.gz under frontend folder"

dist-be: dist-be-linux dist-be-darwin dist-be-windows

# Target with export for Go environment variables
dist-be-linux: export GOOS=linux
dist-be-linux: export GOARCH=$(ARCH64)
dist-be-linux:
	@echo "Building Go backend for Linux"
ifeq ($(OS),Windows_NT)
	cd backend && go build -ldflags "-X main.buildDate=$(DATE)" -o $(DIST_DIR)/$(PACKAGE_NAME)-linux . && cd $(DIST_DIR) && $(TAR) -czf $(PACKAGE_NAME)-linux-$(ARCH64).tar.gz $(PACKAGE_NAME)-linux && $(RM) $(PACKAGE_NAME)-linux
else
	cd backend && go build -ldflags "-X main.buildDate=$(DATE)" -o $(DIST_DIR)/$(PACKAGE_NAME)-linux . && cd $(DIST_DIR) && tar czf $(PACKAGE_NAME)-linux-$(ARCH64).tar.gz $(PACKAGE_NAME)-linux && rm $(PACKAGE_NAME)-linux
endif
	@echo "Created $(PACKAGE_NAME)-linux-$(ARCH64).tar.gz"

dist-be-darwin: export GOOS=darwin
dist-be-darwin: export GOARCH=$(ARCH64)
dist-be-darwin:
	@echo "Building Go backend for macOS (Darwin)"
ifeq ($(OS),Windows_NT)
	cd backend && go build -ldflags "-X main.buildDate=$(DATE)" -o $(DIST_DIR)/$(PACKAGE_NAME)-darwin . && cd $(DIST_DIR) && $(TAR) -czf $(PACKAGE_NAME)-darwin-$(ARCH64).tar.gz $(PACKAGE_NAME)-darwin && $(RM) $(PACKAGE_NAME)-darwin
else
	cd backend && go build -ldflags "-X main.buildDate=$(DATE)" -o $(DIST_DIR)/$(PACKAGE_NAME)-darwin . && cd $(DIST_DIR) && tar czf $(PACKAGE_NAME)-darwin-$(ARCH64).tar.gz $(PACKAGE_NAME)-darwin && rm $(PACKAGE_NAME)-darwin
endif
	@echo "Created $(PACKAGE_NAME)-darwin-$(ARCH64).tar.gz"

dist-be-windows: export GOOS=windows
dist-be-windows: export GOARCH=$(ARCH64)
dist-be-windows:
	@echo "Building Go backend for Windows"
ifeq ($(OS),Windows_NT)
	cd backend && go build -ldflags "-X main.buildDate=$(DATE)" -o $(DIST_DIR)/$(PACKAGE_NAME)-windows.exe . && cd $(DIST_DIR) && powershell -Command "Compress-Archive -Path $(PACKAGE_NAME)-windows.exe -DestinationPath $(PACKAGE_NAME)-windows-$(ARCH64).zip -Force" && $(RM) $(PACKAGE_NAME)-windows.exe
else
	cd backend && go build -ldflags "-X main.buildDate=$(DATE)" -o $(DIST_DIR)/$(PACKAGE_NAME)-windows.exe . && cd $(DIST_DIR) && zip $(PACKAGE_NAME)-windows-$(ARCH64).zip $(PACKAGE_NAME)-windows.exe && rm $(PACKAGE_NAME)-windows.exe
endif
	@echo "Created $(PACKAGE_NAME)-windows-$(ARCH64).zip"

generate-self-signed-cert:
	@echo "Generate self signed certificate for BE"
	cd backend && openssl req -x509 -nodes -newkey rsa:4096 -keyout decrypted_key.pem -out cert-streamer.pem -days 9999 -subj "/C=IT/ST=Italy/L=Turin/O=MyCompany/OU=MyDivision"
