.PHONY: cover
cover:
	go test -covermode=count -coverprofile=main.cov github.com/FreekKalter/km
	go test -covermode=count -coverprofile=lib.cov  github.com/FreekKalter/km/lib
	tail -n +2 lib.cov >> main.cov
	go tool cover -func=main.cov | tail -1
	go tool cover -html=main.cov
