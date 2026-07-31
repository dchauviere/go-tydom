setup:
	@go get
	@(cd $(CURDIR)/pkg/gateway/ui && npm install)

$(CURDIR)/pkg/gateway/ui/dist/index.html:
	@echo "Building Web UI"
	@(cd $(CURDIR)/pkg/gateway/ui && npm run build)

build: $(CURDIR)/pkg/gateway/ui/dist/index.html
	@echo "Building go-tydom"
	@go build

snapshot: $(CURDIR)/pkg/gateway/ui/dist/index.html
	@goreleaser release --snapshot --clean

release: $(CURDIR)/pkg/gateway/ui/dist/index.html
	@goreleaser release --clean

clean:
	@rm -rf $(CURDIR)/pkg/gateway/ui/dist
	@rm -rf $(CURDIR)/pkg/gateway/ui/node_modules
	@rm -f $(CURDIR)/go-tydom
	@rm -rf $(CURDIR)/dist
