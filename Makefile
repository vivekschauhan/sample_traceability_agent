.PHONY: all dep test build 

WORKSPACE ?= $$(pwd)

GO_PKG_LIST := $(shell go list ./... | grep -v /vendor/)

dep:
	@echo "Resolving go package dependencies"
	@go mod tidy
	@echo "Package dependencies completed"

update-sdk:
	@echo "Updating SDK dependencies"
	@export GOFLAGS="-mod=mod" & go mod edit -require "github.com/Axway/agent-sdk@main"


${WORKSPACE}/sample_traceability_agent: dep
	@export time=`date +%Y%m%d%H%M%S` && \
	export CGO_ENABLED=0 && \
	export version=`cat version` && \
	export commit_id=`git rev-parse --short HEAD` && \
	export sdk_version=`go list -m github.com/Axway/agent-sdk | awk '{print $$2}' | awk -F'-' '{print substr($$1, 2)}'` && \
	go build -v -tags static_all \
		-ldflags="-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildTime=$${time}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildVersion=$${version}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildCommitSha=$${commit_id}' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentName=SampleTraceabilityAgent' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.BuildAgentDescription=Sample Traceability Agent' \
				-X 'github.com/Axway/agent-sdk/pkg/cmd.SDKBuildVersion=$${sdk_version}'" \
		-a -o ${WORKSPACE}/bin/sample_traceability_agent ${WORKSPACE}/main.go

build:${WORKSPACE}/sample_traceability_agent
	@echo "Build complete"
