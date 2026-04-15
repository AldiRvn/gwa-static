run:
	docker build -t ts .; docker images ts; docker run -p 8081:8081 ts
