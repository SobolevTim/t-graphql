package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/SobolevTim/t-graphql/internal/config"
	"github.com/SobolevTim/t-graphql/internal/graph/generated"
	"github.com/SobolevTim/t-graphql/internal/graph/resolvers"
	"github.com/SobolevTim/t-graphql/internal/service"
	"github.com/SobolevTim/t-graphql/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Инициализируем хранилище
	store, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("Error initializing store: %v", err)
	}

	r := gin.Default()

	postService := service.NewPostService(store)
	commentService := service.NewCommentService(store)
	subscriptionService := service.NewSubscriptionService(store)

	// Создаём GraphQL-сервер
	resolver := resolvers.NewResolver(postService, commentService, subscriptionService)
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Добавляем транспорты для обработки запросов
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})

	// Логирование запросов
	r.Use(gin.Logger())

	// Регистрируем endpoint для всех HTTP-методов, чтобы WebSocket-запросы (GET) тоже обрабатывались.
	r.Any("/graphql", gin.WrapH(srv))
	r.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))

	log.Println("🚀 GraphQL API running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
