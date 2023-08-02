test:
	go test ./...

bench:
	go test -benchmem -run=^$$ -bench . github.com/lxzan/concurrency/benchmark

cover:
	go test -coverprofile=./bin/cover.out --cover ./...