SHELL = /bin/bash
TARGET = zero

$(TARGET): main.go
	go build -o $(TARGET) main.go

.PHONY: clean
clean:
	rm -f $(TARGET)

.PHONY: run
run: $(TARGET)
	OLLAMA_HOST=http://chiba:11434 ./$(TARGET)
