VERSION       ?= $(shell git describe --tags --always --dirty)
CGO_ENABLED   ?= 0
DIST_DIR      := dist
HOSTLOG_PKG   := ./cmd/hostlog
PLUGINLOG_PKG := ./cmd/pluginlog
TODAYLOG_PKG  := ./cmd/todaylog
DEVFMT_PKG    := ./cmd/devfmt

LDFLAGS_HOSTLOG   := -X 'github.com/andareed/siftly-hostlog/internal/hostlog.Version=$(VERSION)'
LDFLAGS_PLUGINLOG := -X 'github.com/andareed/siftly-hostlog/internal/pluginlog.Version=$(VERSION)'
LDFLAGS_TODAYLOG  := -X 'github.com/andareed/siftly-hostlog/internal/todaylog.Version=$(VERSION)'
LDFLAGS_DEVFMT    := -X 'github.com/andareed/siftly-hostlog/internal/devfmt.Version=$(VERSION)'

.PHONY: all clean release \
	linux linux-hostlog linux-pluginlog linux-todaylog linux-devfmt \
	windows windows-hostlog windows-pluginlog windows-todaylog windows-devfmt \
	mac mac-amd64 mac-arm64 \
	mac-amd64-hostlog mac-amd64-pluginlog mac-amd64-todaylog mac-amd64-devfmt \
	mac-arm64-hostlog mac-arm64-pluginlog mac-arm64-todaylog mac-arm64-devfmt

all: linux windows mac

clean:
	rm -rf $(DIST_DIR)

linux: linux-hostlog linux-pluginlog linux-todaylog linux-devfmt

linux-hostlog:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_HOSTLOG)" \
		-o $(DIST_DIR)/hostlog_$(VERSION)_linux_amd64 $(HOSTLOG_PKG)

linux-pluginlog:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_PLUGINLOG)" \
		-o $(DIST_DIR)/pluginlog_$(VERSION)_linux_amd64 $(PLUGINLOG_PKG)

linux-todaylog:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_TODAYLOG)" \
		-o $(DIST_DIR)/todaylog_$(VERSION)_linux_amd64 $(TODAYLOG_PKG)

linux-devfmt:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_DEVFMT)" \
		-o $(DIST_DIR)/devfmt_$(VERSION)_linux_amd64 $(DEVFMT_PKG)

windows: windows-hostlog windows-pluginlog windows-todaylog windows-devfmt

windows-hostlog:
	mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS_HOSTLOG)" \
		-o $(DIST_DIR)/hostlog_$(VERSION)_windows_amd64.exe $(HOSTLOG_PKG)

windows-pluginlog:
	mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS_PLUGINLOG)" \
		-o $(DIST_DIR)/pluginlog_$(VERSION)_windows_amd64.exe $(PLUGINLOG_PKG)

windows-todaylog:
	mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS_TODAYLOG)" \
		-o $(DIST_DIR)/todaylog_$(VERSION)_windows_amd64.exe $(TODAYLOG_PKG)

windows-devfmt:
	mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS_DEVFMT)" \
		-o $(DIST_DIR)/devfmt_$(VERSION)_windows_amd64.exe $(DEVFMT_PKG)

mac: mac-amd64 mac-arm64

mac-amd64: mac-amd64-hostlog mac-amd64-pluginlog mac-amd64-todaylog mac-amd64-devfmt

mac-amd64-hostlog:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS_HOSTLOG)" \
		-o $(DIST_DIR)/hostlog_$(VERSION)_darwin_amd64 $(HOSTLOG_PKG)

mac-amd64-pluginlog:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS_PLUGINLOG)" \
		-o $(DIST_DIR)/pluginlog_$(VERSION)_darwin_amd64 $(PLUGINLOG_PKG)

mac-amd64-todaylog:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS_TODAYLOG)" \
		-o $(DIST_DIR)/todaylog_$(VERSION)_darwin_amd64 $(TODAYLOG_PKG)

mac-amd64-devfmt:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS_DEVFMT)" \
		-o $(DIST_DIR)/devfmt_$(VERSION)_darwin_amd64 $(DEVFMT_PKG)

mac-arm64: mac-arm64-hostlog mac-arm64-pluginlog mac-arm64-todaylog mac-arm64-devfmt

mac-arm64-hostlog:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS_HOSTLOG)" \
		-o $(DIST_DIR)/hostlog_$(VERSION)_darwin_arm64 $(HOSTLOG_PKG)

mac-arm64-pluginlog:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS_PLUGINLOG)" \
		-o $(DIST_DIR)/pluginlog_$(VERSION)_darwin_arm64 $(PLUGINLOG_PKG)

mac-arm64-todaylog:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS_TODAYLOG)" \
		-o $(DIST_DIR)/todaylog_$(VERSION)_darwin_arm64 $(TODAYLOG_PKG)

mac-arm64-devfmt:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS_DEVFMT)" \
		-o $(DIST_DIR)/devfmt_$(VERSION)_darwin_arm64 $(DEVFMT_PKG)

release: clean all
	@echo "Built release binaries in $(DIST_DIR):"
	@ls -1 $(DIST_DIR)
