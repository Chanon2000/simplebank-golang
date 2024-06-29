DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
# DB_URL=postgresql://root:TK3NVg8y63kTrdJsZqg5@simple-bank.c7btprkzoik4.ap-southeast-1.rds.amazonaws.com:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/chanon2000/simplebank/db/sqlc Store

db_docs: # เพื่อ re-generate doc
	dbdocs build doc/db.dbml

db_schema: # เพื่อ re-generate schema sql code
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

proto: # เอามาจาก doc ของ proto แล้วเอามา update เพิ่มเติมอีกที
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	proto/*.proto
# rm -f pb/*.go เพื่อลบ .go files ใน pb folder ออกให้หมดก่อน regenerate (เพื่อบางครั้งเราลบ proto files เมื่อ regenerate .go ที่ได้จาก .proto file นั้นจะได้หายไป เพื่อให้ code clean ขึ้นนั้นเอง)
# --proto_path เพื่อ point ไปที่ proto directory
# --go_out เพื่อ point ไปที่ที่ generated golang code จะวาง
# --go-grpc_out คือ point qrpc output
# proto/*.proto คือ location ของ proto files โดย proto/*.proto หมายถึง .proto files ทั้งหมดใน proto folder
# --grpc-gateway... คือ เพิ่ม generate grpc gateway
#  ซึ่งเมื่อ generate ก็จะเห็น .pb.gw.go file ซึ่งคือ generated gateway code นั้น located อยู่

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock db_docs db_schema proto evans