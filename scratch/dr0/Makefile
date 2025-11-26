SHELL = /bin/bash
TARGET = dr0

$(TARGET): main.go
	go build -o $(TARGET) main.go

.PHONY: clean
clean:
	rm -f $(TARGET)

.PHONY: run
run: $(TARGET)
	OLLAMA_HOST=http://k9:11434 ./$(TARGET)
