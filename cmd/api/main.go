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
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Устанавливаем доверенные прокси-сервера
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.1"})

	// Создаем сервисы для работы с постами и комментариями
	postService := service.NewPostService(store)
	commentService := service.NewCommentService(store)
	subscriptionService := service.NewSubscriptionService(store)

	// Создаём резолверы
	resolver := resolvers.NewResolver(postService, commentService, subscriptionService)
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Добавляем транспорты для обработки GraphQL запросов
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})

	// Регистрируем эндпоинт для GraphQL (POST и WebSocket запросы)
	r.Any("/graphql", gin.WrapH(srv))

	// Добавляем Playground для тестирования запросов
	r.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))

	// Запуск API-сервера
	log.Println("GraphQL API running at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
