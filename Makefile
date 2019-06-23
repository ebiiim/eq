ROOT=github.com/ebiiim/eq
BIN=bin/eq
MAIN=main.go
TEST=...
COVERAGE_FILE=cover

CMD_ALL=all
CMD_BUILD=build
CMD_TEST=test
CMD_COVERAGE=coverage

.PHONY: ${CMD_ALL}
${CMD_ALL}: ${CMD_TEST} ${CMD_BUILD}

.PHONY: ${CMD_BUILD}
${CMD_BUILD}: ${MAIN}
	go build -o ${BIN} ${GOPATH}/src/${ROOT}/$?

.PHONY: ${CMD_TEST}
${CMD_TEST}:
	go test -race -cover ${ROOT}/${TEST}

.PHONY: ${CMD_COVERAGE}
${CMD_COVERAGE}:
	go test -race -coverprofile=${COVERAGE_FILE}.out ${ROOT}/${TEST}
	go tool cover -func=${COVERAGE_FILE}.out
	go tool cover -html=${COVERAGE_FILE}.out -o ${COVERAGE_FILE}.html