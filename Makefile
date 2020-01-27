.PHONY: build
build: clean bin/cm500_exporter
clean:
	rm -rf bin
	
bin/cm500_exporter: $(shell find . -type f -name '*.go' | grep -v '_test.go')
	@mkdir bin
	go build -o bin/cm500_exporter .