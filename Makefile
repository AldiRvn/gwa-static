run-win:
	go build -v -o gen/app.exe .
	gen/app.exe

run:
	docker build -t ts .; docker images ts; docker rm -f ts; docker run -p 8081:8081 --network host --name ts ts

repo:
	cmd.exe /c start https://github.com/AldiRvn/gwa-static

update:
	go mod tidy; go get -u; go mod vendor
