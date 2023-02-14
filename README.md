# todo with htmx spike

spiking out a todo list with htmx. 

all the fluid UI fun of an SPA with none of the pain

![image](https://user-images.githubusercontent.com/631756/205446910-2196c5e5-ffe5-418d-b468-9523d0d2d954.png)

## try it

`go run cmd/server/main.go` and [visit](http://localhost:8000)

## notes, thoughts etc

[see the twitter thread for videos, thoughts, etc](https://twitter.com/quii/status/1598987894865113088)

```shell
# golang migrate with sqlite
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
# sqlc sqlite support required
go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.16.0

sqlc compile
sqlc generate

migrate create -dir domain/sql/migration/ -ext sql init
...
migrate -path domain/sql/migration/ -database sqlite:///tmp/test.db up
```


CGO_ENABLED=0 go build -o todo cmd/server/main.go