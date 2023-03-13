gen:
# protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb
	protoc --go_out=.  --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
rm:
	rm -rf pb
run:
	go run main.go