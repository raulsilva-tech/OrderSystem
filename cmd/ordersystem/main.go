package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/raulsilva-tech/OrderSystem/configs"
	"github.com/raulsilva-tech/OrderSystem/internal/event/handler"
	"github.com/raulsilva-tech/OrderSystem/internal/infra/graph"
	"github.com/raulsilva-tech/OrderSystem/internal/infra/grpc/pb"
	"github.com/raulsilva-tech/OrderSystem/internal/infra/grpc/service"
	"github.com/raulsilva-tech/OrderSystem/internal/infra/web/webserver"
	"github.com/raulsilva-tech/OrderSystem/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	//getting environment variables
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	//creating database connection
	db, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//getrabbitMQChannel and register a new event
	rabbitMQChannel := getRabbitMQChannel()
	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	//creating use cases
	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	listOrdersUseCase := NewListOrdersUseCase(db)

	// creating a new gRPC server
	grpcServer := grpc.NewServer()
	orderService := service.NewOrderService(*createOrderUseCase, *listOrdersUseCase)
	pb.RegisterOrderServiceServer(grpcServer, orderService)

	reflection.Register(grpcServer)
	fmt.Println("Starting gRPC server on port", configs.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.GRPCServerPort))
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(lis)

	//creating GraphQL server
	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
		ListOrdersUseCase:  *listOrdersUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)
	fmt.Println("Starting GraphQL server on port", configs.GraphQLServerPort)
	go http.ListenAndServe(":"+configs.GraphQLServerPort, nil)

	//creating webserver
	webserver := webserver.NewWebServer(configs.WebServerPort)
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	webserver.AddHandler("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(4 * time.Second)
		w.Write([]byte("Webserver running successfully \n"))
	})
	webserver.AddHandler("/order", webOrderHandler.Create)
	webserver.AddHandler("/orders", webOrderHandler.GetAll)
	fmt.Println("Starting web server on port", configs.WebServerPort)
	go webserver.Start()

	//create channel to receive system interruptions
	stop := make(chan os.Signal, 1)
	//set the channel to receive the signal
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	//wait for the signal
	<-stop

	//set a new context with 5 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Shutting down web server ... ")
	//shutting down the server using the context with timeout interval
	if err := webserver.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdwon the server: %v \n", err)
	}

	fmt.Println("Server stopped")
}

func getRabbitMQChannel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}
