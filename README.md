- gRPC server on port 50051

- Web server on port 8000

- GraphQL server on port 8080

- To create the database tables:
 migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/orders" -verbose up
