
.PHONY: all
all: test_build test cli

.PHONY: cli
cli:
	cd cmd/cli && go build -o po cli.go && mv po ${GOPATH}/bin/po

.PHONY: test_build
test_build:
	cd cmd/cli && go build cli.go && rm cli
	cd cmd/api && go build main.go && rm main
	cd cmd/cdn && go build main.go && rm main
	
.PHONY: test
test:
	go test
	cd internal && go test
	cd internal/cdn && go test
	cd internal/metadata && go test
	cd auth && go test
	cd client && go test
	cd config && go test
	cd feed && go test
	
.PHONY: test_coverage
test_coverage:
	go test `go list ./... | grep -v cmd` -coverprofile=coverage.txt -covermode=atomic

.PHONY: prepare_cdn
prepare_cdn:
	cd deployment && ansible-playbook -i inventory/podops.dev.yml playbooks/prepare_cdn.yml

.PHONY: deploy_cdn
deploy_cdn:
	cd deployment && ansible-playbook -i inventory/podops.dev.yml playbooks/deploy_containers.yml