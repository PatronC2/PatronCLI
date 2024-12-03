build:
	docker build -t go-app-builder .
	docker run --rm -v "$(pwd)/output":/output go-app-builder cp /root/main /output/patron

