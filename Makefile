
OUTPUT_FILE=oauth2svc
PROJECT=oauth2svc
TARGET=./



all: clean vet lint linux

native:
	go build -v -o $(OUTPUT_FILE) $(TARGET)

linux:
	GOOS=linux GOARCH=amd64 go build -v -o $(OUTPUT_FILE) $(TARGET)


vet:
	go vet ./


lint:
	golint ./

clean:
	-rm -f $(OUTPUT_FILE)

