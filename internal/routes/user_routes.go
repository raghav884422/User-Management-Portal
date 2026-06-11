package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/user-api/internal/handler"
)

// RegisterUserRoutes sets up all /users routes on the given Fiber app.
func RegisterUserRoutes(app *fiber.App, h *handler.UserHandler) {
	users := app.Group("/users")

	// POST /users        → create a new user
	users.Post("/", h.CreateUser)

	// GET /users         → list all users (with pagination)
	users.Get("/", h.ListUsers)

	// GET /users/:id     → get a user by ID (includes calculated age)
	users.Get("/:id", h.GetUserByID)

	// PUT /users/:id     → update a user
	users.Put("/:id", h.UpdateUser)

	// DELETE /users/:id  → delete a user
	users.Delete("/:id", h.DeleteUser)
}
