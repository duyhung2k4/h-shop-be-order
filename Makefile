gen_grpc_protoc:
	protoc \
	--go_out=grpc \
	--go_opt=paths=source_relative \
	--go-grpc_out=grpc \
	--go-grpc_opt=paths=source_relative \
	proto/*.proto
export_path:
	export PATH=$PATH:$(go env GOPATH)/bin
gen_key:
	openssl \
		req -x509 \
		-nodes \
		-days 365 \
		-newkey rsa:2048 \
		-keyout keys/server-order/private.pem \
		-out keys/server-order/public.pem \
		-config keys/server-order/san.cfg

test_request_order:
	ab \
    -n 1000 \
    -c 100 \
    -T application/json \
    -p test/order.json \
    -m POST \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImJhYmEwMzk4MjYyMTI0QGdtYWlsLmNvbSIsImV4cCI6MTcxMzM4NTMwMiwicHJvZmlsZV9pZCI6Nywicm9sZSI6InVzZXIiLCJzdWIiOiIxMTc5OTY0MjMxODM5MDc1NzAyNTIiLCJ1dWlkIjoiOWU3ZWY5NWQtZGNhMS00ZTJkLWI1ODktYjYyMjJmM2ZmMDAzIn0.Gc0nuMuxfDU-deS2Y1To0icsJ5C9RkaayLRjhCCW4EU" \
    http://localhost:18886/order/api/v1/protected/order
