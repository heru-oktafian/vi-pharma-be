package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"pharma-be/internal/delivery/http/handler"
	"pharma-be/internal/delivery/http/middleware"
	"pharma-be/internal/delivery/http/route"
	ws "pharma-be/internal/delivery/http/websocket"
	"pharma-be/internal/repository/memory"
	"pharma-be/internal/repository/postgres"
	"pharma-be/internal/usecase"
	"pharma-be/pkg/config"
)

func main() {
	log.Println("Starting Vi-Pharma Backend Service...")

	// 1. Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application configuration: %v", err)
	}

	// 2. Initialize Database & Run Auto-Migration
	db, err := postgres.InitDB(cfg)
	if err != nil {
		log.Printf("[WARNING] Database connection failed: %v", err)
		log.Println("[INFO] Please verify your PostgreSQL server is running and .env credentials are correct.")
	} else {
		log.Println("[DATABASE] PostgreSQL connected & schema migrated successfully.")
	}

	// 3. Initialize Repositories & In-Memory Stores
	userRepo := postgres.NewUserRepository(db)
	branchRepo := postgres.NewBranchRepository(db)
	qrSessionStore := memory.NewQRSessionStore()
	sessionStore := memory.NewSessionStore()

	unitRepo := postgres.NewUnitRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	supplierRepo := postgres.NewSupplierRepository(db)
	doctorRepo := postgres.NewDoctorRepository(db)
	medicineRepo := postgres.NewMedicineRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	procurementRepo := postgres.NewProcurementRepository(db)
	posRepo := postgres.NewPOSRepository(db)

	// 4. Initialize WebSocket Hub
	wsHub := ws.NewHub()

	// 5. Initialize Clean Architecture Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg)
	qrAuthUsecase := usecase.NewQRAuthUsecase(qrSessionStore, sessionStore, userRepo, branchRepo, wsHub, cfg)
	userUsecase := usecase.NewUserUsecase(userRepo)
	branchUsecase := usecase.NewBranchUsecase(branchRepo)

	unitUsecase := usecase.NewUnitUsecase(unitRepo)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	supplierUsecase := usecase.NewSupplierUsecase(supplierRepo)
	doctorUsecase := usecase.NewDoctorUsecase(doctorRepo)
	medicineUsecase := usecase.NewMedicineUsecase(medicineRepo, categoryRepo, unitRepo)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, medicineRepo)
	procurementUsecase := usecase.NewProcurementUsecase(procurementRepo, supplierRepo, medicineRepo)
	posUsecase := usecase.NewPOSUsecase(posRepo, db)

	// 6. Initialize HTTP & WebSocket Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	qrAuthHandler := handler.NewQRAuthHandler(qrAuthUsecase, sessionStore, wsHub, cfg)
	userHandler := handler.NewUserHandler(userUsecase)
	branchHandler := handler.NewBranchHandler(branchUsecase)

	unitHandler := handler.NewUnitHandler(unitUsecase)
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	supplierHandler := handler.NewSupplierHandler(supplierUsecase)
	doctorHandler := handler.NewDoctorHandler(doctorUsecase)
	medicineHandler := handler.NewMedicineHandler(medicineUsecase)
	inventoryHandler := handler.NewInventoryHandler(inventoryUsecase)
	procurementHandler := handler.NewProcurementHandler(procurementUsecase)
	posHandler := handler.NewPOSHandler(posUsecase)

	// 7. Initialize Fiber Web App
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ServerHeader: "Fiber",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return handler.ErrorResponse(c, code, err.Error(), nil)
		},
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))
	app.Use(middleware.RequestLogger())

	// 8. Setup Routes
	route.SetupRoutes(
		app,
		cfg,
		sessionStore,
		authHandler,
		qrAuthHandler,
		userHandler,
		branchHandler,
		unitHandler,
		categoryHandler,
		supplierHandler,
		doctorHandler,
		medicineHandler,
		inventoryHandler,
		procurementHandler,
		posHandler,
	)

	// 9. Graceful Shutdown Handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down Vi-Pharma server...")
		if err := app.Shutdown(); err != nil {
			log.Fatalf("Error shutting down server: %v", err)
		}
	}()

	// 10. Start Server
	serverAddr := ":" + cfg.AppPort
	log.Printf("Vi-Pharma Backend running on port %s (env: %s)\n", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(serverAddr); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}
