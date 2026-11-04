package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"pharma-be/internal/delivery/http/handler"
	"pharma-be/internal/delivery/http/middleware"
	"pharma-be/internal/domain"
	"pharma-be/pkg/config"
)

// SetupRoutes registers all application endpoints on the Fiber app
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	sessionStore domain.SessionStore,
	authHandler *handler.AuthHandler,
	qrAuthHandler *handler.QRAuthHandler,
	userHandler *handler.UserHandler,
	branchHandler *handler.BranchHandler,
	unitHandler *handler.UnitHandler,
	categoryHandler *handler.CategoryHandler,
	supplierHandler *handler.SupplierHandler,
	doctorHandler *handler.DoctorHandler,
	medicineHandler *handler.MedicineHandler,
) {
	jwtAuth := middleware.JWTProtected(cfg, sessionStore)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// 1. Health check endpoint
	v1.Get("/health", func(c *fiber.Ctx) error {
		return handler.SuccessResponse(c, fiber.StatusOK, "Vi-Pharma Backend Service is healthy", fiber.Map{
			"app":     cfg.AppName,
			"version": "1.0.0",
			"status":  "online",
		})
	})

	// 2. Authentication & QR Pairing Group
	auth := v1.Group("/auth")

	// 2.1 Standard Mobile Login (vi-token uses this to get mobile JWT)
	auth.Post("/login", authHandler.Login)

	// 2.2 QR Code Flow (vi-pharma web pairing with vi-token mobile)
	qr := auth.Group("/qr")
	qr.Post("/generate", qrAuthHandler.GenerateQR)                  // Web requests QR
	qr.Get("/status/:sessionId", qrAuthHandler.GetSessionStatus)    // Polling fallback
	qr.Post("/select-branch", qrAuthHandler.SelectBranch)           // Web chooses branch

	// 2.3 WebSocket upgrade middleware & endpoint for vi-pharma web QR pairing
	qr.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	qr.Get("/ws/:sessionId", qrAuthHandler.HandleWebSocket())

	// 2.4 Mobile QR Authorization (Protected by Mobile JWT)
	qrProtected := qr.Group("")
	qrProtected.Use(jwtAuth)
	qrProtected.Post("/authorize", qrAuthHandler.AuthorizeQR)

	// 2.5 Profile Info & Remote Logout Actions (Protected)
	authProtected := auth.Group("")
	authProtected.Use(jwtAuth)
	authProtected.Get("/me", authHandler.GetMe)
	authProtected.Post("/remote-logout", qrAuthHandler.RemoteLogout)
	authProtected.Get("/active-workstation", qrAuthHandler.GetActiveWorkstation)

	// 3. User Management (Protected: SuperAdmin / Admin)
	users := v1.Group("/users")
	users.Use(jwtAuth)
	users.Use(middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin))
	users.Post("", userHandler.CreateUser)
	users.Get("", userHandler.ListUsers)
	users.Get("/:id", userHandler.GetUserByID)

	// 4. Branch Management (Protected: SuperAdmin / Admin)
	branches := v1.Group("/branches")
	branches.Use(jwtAuth)
	branches.Use(middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin))
	branches.Post("", branchHandler.CreateBranch)
	branches.Get("", branchHandler.ListBranches)
	branches.Get("/:id", branchHandler.GetBranchByID)
	branches.Post("/assign", branchHandler.AssignUserBranch)

	// 5. Workstation / Active Branch Features (Protected by Selected Branch)
	// 5.1 Workstation WebSocket listener for real-time remote logout
	v1.Use("/workstation/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	v1.Get("/workstation/ws", qrAuthHandler.HandleWorkstationWebSocket())

	workstation := v1.Group("/workstation")
	workstation.Use(jwtAuth)
	workstation.Use(middleware.RequireBranch())

	workstation.Get("/current", func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		role := c.Locals("role")
		branchID := c.Locals("branchID")

		return handler.SuccessResponse(c, fiber.StatusOK, "Active branch session verified", fiber.Map{
			"user_id":   userID,
			"role":      role,
			"branch_id": branchID,
		})
	})

	// -------------------------------------------------------------
	// 6. Master Data Farmasi (Protected by JWT)
	// -------------------------------------------------------------

	// 6.1 Units (Satuan)
	units := v1.Group("/units")
	units.Use(jwtAuth)
	units.Get("", unitHandler.ListUnits)
	units.Get("/:id", unitHandler.GetUnitByID)
	units.Post("", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), unitHandler.CreateUnit)
	units.Put("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), unitHandler.UpdateUnit)
	units.Delete("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin), unitHandler.DeleteUnit)

	// 6.2 Categories (Kategori & Golongan Terapi)
	categories := v1.Group("/categories")
	categories.Use(jwtAuth)
	categories.Get("", categoryHandler.ListCategories)
	categories.Get("/:id", categoryHandler.GetCategoryByID)
	categories.Post("", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), categoryHandler.CreateCategory)
	categories.Put("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), categoryHandler.UpdateCategory)
	categories.Delete("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin), categoryHandler.DeleteCategory)

	// 6.3 Suppliers / PBF (Pedagang Besar Farmasi)
	suppliers := v1.Group("/suppliers")
	suppliers.Use(jwtAuth)
	suppliers.Get("", supplierHandler.ListSuppliers)
	suppliers.Get("/:id", supplierHandler.GetSupplierByID)
	suppliers.Post("", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), supplierHandler.CreateSupplier)
	suppliers.Put("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), supplierHandler.UpdateSupplier)
	suppliers.Delete("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin), supplierHandler.DeleteSupplier)

	// 6.4 Doctors (Dokter & Tenaga Medis)
	doctors := v1.Group("/doctors")
	doctors.Use(jwtAuth)
	doctors.Get("", doctorHandler.ListDoctors)
	doctors.Get("/:id", doctorHandler.GetDoctorByID)
	doctors.Post("", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), doctorHandler.CreateDoctor)
	doctors.Put("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), doctorHandler.UpdateDoctor)
	doctors.Delete("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin), doctorHandler.DeleteDoctor)

	// 6.5 Medicines & Medical Supplies (Katalog Master Obat & Alkes)
	medicines := v1.Group("/medicines")
	medicines.Use(jwtAuth)
	medicines.Get("", medicineHandler.ListMedicines)
	medicines.Get("/barcode/:barcode", medicineHandler.GetMedicineByBarcode)
	medicines.Get("/:id", medicineHandler.GetMedicineByID)
	medicines.Post("", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), medicineHandler.CreateMedicine)
	medicines.Put("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist), medicineHandler.UpdateMedicine)
	medicines.Delete("/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin), medicineHandler.DeleteMedicine)
}
