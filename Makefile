run:
	docker build -t ts .; docker images ts; docker rm -f ts; docker run -p 8081:8081 --name ts ts

update:
	go mod tidy; go get -u; go mod vendor
