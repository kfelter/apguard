build:
	docker build -f ./cmd/rps/Dockerfile . -t apguard/rpx:latest
	docker build -f ./cmd/greeter/Dockerfile . -t apguard/greeter:latest

push:
	docker push apguard/rpx:latest
	docker push apguard/greeter:latest

ex:
	docker-compose -f example/docker-compose.yml up -d 
	curl localhost:9001
	curl localhost:9001

