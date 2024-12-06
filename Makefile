build:
	docker build --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t go-app-builder .
ifeq ($(OS),Windows_NT)
	docker run --rm -v "%cd%/output:/output" go-app-builder cp /root/main /output/patron.exe
else
	docker run --rm -v "$(pwd)/output":/output go-app-builder cp /root/main /output/patron
endif
