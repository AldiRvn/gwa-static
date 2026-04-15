FROM golang:1.26-alpine AS build

COPY . .

# CGO_ENABLED=0 go build -o /app .
#           IMAGE       ID             DISK USAGE   CONTENT SIZE   EXTRA
# results = ts:latest   54d373c60c5e         85MB         22.9MB    U   
#? Reduce binary size for deployment production.
#? reference: https://go-cookbook.com/snippets/builds-and-compilations/optimizing-build-sizes#:~:text=go%20build%20%2Dtrimpath%20%2Dldflags%20%22%2Dw%20%2Ds%22%20microservice.go
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-w -s" -o /app .
#           IMAGE       ID             DISK USAGE   CONTENT SIZE   EXTRA
# results = ts:latest   b2a5b3bf0c2c       67.5MB         15.5MB

FROM scratch

COPY --from=build /app /app

#? Support HTTPS Request
#? reference: https://gist.github.com/michaelboke/564bf96f7331f35f1716b59984befc50#:~:text=builder%20/app/app%20.-,COPY%20%2D%2Dfrom%3Dbuilder%20/etc/ssl/certs/ca%2Dcertificates.crt%20/etc/ssl/certs/,-CMD%20%5B%22./app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#? Don't forget to add import _ "time/tzdata" in main.go so this env will work.
ENV TZ=Asia/Jakarta

ENTRYPOINT [ "/app" ]
