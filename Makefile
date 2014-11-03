.PHONY: cover
cover:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONE: heatmap
heatmap:
	go test -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out
