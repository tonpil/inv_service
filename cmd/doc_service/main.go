package main

import (
	"doc_service/config"
	cache "doc_service/internal/infra/cache"
	"doc_service/internal/infra/http"
	db "doc_service/internal/infra/reindexer"
	"doc_service/internal/infra/service"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config", err)
	}

	dbClient, err := db.NewClient(
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.User,
		cfg.DB.Password,
	)
	if err != nil {
		log.Fatal("failed to connect to db", err)
	}
	defer dbClient.Close()

	err = db.InitNamespaces(dbClient)
	if err != nil {
		log.Fatal("failed to init namespaces", err)
	}

	docRepository := db.NewRepository(dbClient)
	docCache := cache.NewInMemoryCache()
	unitOfWork := db.NewUnitOfWork(dbClient)

	docService := service.NewDocumentService(docRepository, unitOfWork, docCache)

	docHandler := http.NewDocumentHandler(docService)
	router := http.SetupRouter(docHandler)
	server := http.NewServer(router, cfg.Service.Host, cfg.Service.Port)

	server.Run()
}
