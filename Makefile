VERSION=0.0.1
IMAGE_NAME=hello

IMAGE=$(IMAGE_NAME)-$(VERSION).img
all: $(IMAGE)

$(IMAGE): distroless.layer hello.layer
	ocibuild image build --tag $(IMAGE_NAME):$(VERSION) --base distroless.layer hello.layer > $(IMAGE)

distroless.layer:
	crane pull gcr.io/distroless/static:nonroot distroless.layer

hello.layer: hello.go go.sum
	ocibuild layer gobuild . > hello.layer

go.sum: go.mod hello.go
	go mod tidy

clean:
	rm -f hello.layer

clobber: clean
	rm -f *.img go.sum distroless.layer

run: $(IMAGE)
	docker load < $(IMAGE)
	docker run -p8080:8080 -it --entrypoint /usr/local/bin/oci-hello $(IMAGE_NAME):$(VERSION)

