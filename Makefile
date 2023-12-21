run:
	go run ./cmd/main.go
build:
	docker build -t forum-image:local .
docker-run:
	docker run --rm -p 8080:8080 -t forum-image:local
docker-stop:
	docker stop $$(forum-image:local)
docker-delete:
	docker rmi $$(forum-image:local) && docker rmi $$(forum-image:local)
