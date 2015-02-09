build:
	go build -o vlcinterface

run:
	./vlcinterface

clean:
	rm vlcinterface

.PHONY: build run clean
