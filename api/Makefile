# Go parameters
    GOCMD=go
    LDFLAGS = -ldflags "-X main.VERSION=${VERSION}"
    GOBUILD=$(GOCMD) build $(LDFLAGS)
    GOCLEAN=$(GOCMD) clean
    GOTEST=$(GOCMD) test
    GOGET=$(GOCMD) get
    BINARY=hackhub
    ORG=thatotherguydev
    VERSION=0.0.2
    DOCKER_PROD_TAG=$(ORG)/$(BINARY):$(VERSION)
    DOCKER_LATEST_TAG=$(ORG)/$(BINARY):latest

    all: build-linux
    build:
	$(GOBUILD) -o $(BINARY_NAME) -v
    test:
	$(GOTEST) -v ./... -cover
    test-html:
	$(GOTEST) -v ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out
    clean:
	$(GOCLEAN)
	rm -f $(BINARY)
	rm coverage.out
    mongo:
	docker volume create mongodbdata
	docker run -p 27017:27017 -v mongodbdata:/data/db -d mongo
	# docker run -p 27017:27017 --mount type=bind,source=mongodbdata,target=/data/db -d mongo
    run:
	$(GOBUILD) -o $(BINARY) -v main.go
	./$(BINARY)
    deps:
	dep ensure

    # Cross compilation
    build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY) -v
    docker-build:
	docker build . -t $(DOCKER_PROD_TAG)
    docker-push: docker-build
	docker push $(DOCKER_PROD_TAG)
	docker tag  $(DOCKER_PROD_TAG) $(DOCKER_LATEST_TAG)
	docker push $(DOCKER_LATEST_TAG)
