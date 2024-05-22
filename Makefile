PROTO_DIR = proto
PACKAGE = $(shell head -1 go.mod | awk '{print $$2}')

protoc:
	rm -rf pb
	protoc -I${PROTO_DIR} --go-grpc_out=. --go_out=. ${PROTO_DIR}/*.proto

migration-%:
	migrate create -ext sql -dir database/migrations $*

migration-up:
	migrate -path ./database/migrations -database "postgresql://chat:chat@localhost:5433/chat?sslmode=disable" up

mockery:
	mockery --all


cert:
	cd cert; ./gen.sh; cd ..