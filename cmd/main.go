package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5"
	"github.com/toxanetoxa/todo-list/internal/handlers"
	"os"
)

func main() {
	app := fiber.New()

	db, err := dbConnect()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close(context.Background())

	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "America/New_York",
		Output:     os.Stdout,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/tasks", handlers.GetTasks(db))

	app.Get("/tasks/:id", handlers.GetTaskById(db))

	app.Post("/tasks", handlers.CreateTask(db))

	app.Delete("/tasks/:id", handlers.DeleteTask(db))

	app.Put("/tasks/:id", handlers.UpdateTask(db))

	PORT := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	err = app.Listen(PORT)
	if err != nil {
		fmt.Print("Server not started \n", err)
	}
}

func dbConnect() (*pgx.Conn, error) {
	dsn := os.Getenv("DB_URI")

	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	fmt.Println("Connected to database")
	return conn, nil
}
