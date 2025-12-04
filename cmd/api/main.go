package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/configs"
	"user-service/internal/domain"
	"user-service/internal/infrastructure/jwt"
	"user-service/internal/infrastructure/logger"
	"user-service/internal/infrastructure/pld"
	"user-service/internal/infrastructure/rabbitmq"
	"user-service/internal/infrastructure/repository"
	httphandler "user-service/internal/interfaces/http"
	"user-service/internal/interfaces/http/handlers"
	"user-service/internal/usecase"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Error al cargar configuración: %v", err)
	}

	appLogger, err := logger.NewLogger(os.Getenv("ENV"))
	if err != nil {
		log.Fatalf("Error al inicializar logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Iniciando aplicación", zap.String("version", "1.0.0"))

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		appLogger.Fatal("Error al conectar a la base de datos", zap.Error(err))
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.UserEvent{}); err != nil {
		appLogger.Fatal("Error al migrar base de datos", zap.Error(err))
	}
	appLogger.Info("Base de datos migrada correctamente")

	userRepo := repository.NewUserRepository(db)
	userEventRepo := repository.NewUserEventRepository(db)

	jwtService := jwt.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.ExpiresIn)
	pldService := pld.NewPLDClient(cfg.PLD.BaseURL, cfg.PLD.Timeout, appLogger)

	eventPublisher, err := rabbitmq.NewEventPublisher(cfg.RabbitMQ.URL, "user.created")
	if err != nil {
		appLogger.Fatal("Error al inicializar publisher de RabbitMQ", zap.Error(err))
	}
	appLogger.Info("Publisher de RabbitMQ inicializado")

	eventConsumer, err := rabbitmq.NewEventConsumer(cfg.RabbitMQ.URL, "user.created")
	if err != nil {
		appLogger.Fatal("Error al inicializar consumer de RabbitMQ", zap.Error(err))
	}
	appLogger.Info("Consumer de RabbitMQ inicializado")

	createUserUseCase := usecase.NewCreateUserUseCase(
		userRepo,
		pldService,
		eventPublisher,
		jwtService,
	)

	loginUseCase := usecase.NewLoginUseCase(
		userRepo,
		jwtService,
	)

	getUserUseCase := usecase.NewGetUserUseCase(userRepo)

	userHandler := handlers.NewUserHandler(
		createUserUseCase,
		loginUseCase,
		getUserUseCase,
	)

	router := httphandler.SetupRouter(userHandler, jwtService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		appLogger.Info("Iniciando consumidor de eventos")
		handler := func(userID, email string, createdAt int64) error {
			appLogger.Info("Procesando evento user.created",
				zap.String("user_id", userID),
				zap.String("email", email),
				zap.Int64("created_at", createdAt),
			)
			
			appLogger.Info("Enviando email de bienvenida",
				zap.String("user_id", userID),
				zap.String("email", email),
			)

			eventPayload, _ := json.Marshal(map[string]interface{}{
				"user_id":    userID,
				"email":      email,
				"created_at": time.Unix(createdAt, 0).Format(time.RFC3339),
			})

			userUUID, err := uuid.Parse(userID)
			if err != nil {
				appLogger.Warn("Error al parsear UUID", zap.Error(err))
				return nil
			}

			event := &domain.UserEvent{
				UserID:    userUUID,
				EventType: "user.created",
				Payload:   eventPayload,
			}

			if err := userEventRepo.Create(context.Background(), event); err != nil {
				appLogger.Warn("Error al guardar evento en auditoría", zap.Error(err))
			}

			return nil
		}

		if err := eventConsumer.ConsumeUserCreated(ctx, handler); err != nil {
			appLogger.Error("Error en consumidor de eventos", zap.Error(err))
		}
	}()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("Servidor HTTP iniciado", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Error al iniciar servidor", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Cerrando servidor...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Error al cerrar servidor", zap.Error(err))
	}

	appLogger.Info("Servidor cerrado")
}

