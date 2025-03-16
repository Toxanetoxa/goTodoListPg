package handlers

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/toxanetoxa/todo-list/internal/models"
	"time"
)

func GetTasks(db *pgx.Conn) fiber.Handler {
	const op = "handlers.task.GetTasks"
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(context.Background(), "SELECT id, title, description, status, created_at, updated_at FROM tasks")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("%s: Не удалось получить данные: %v", op, err)})
		}
		defer rows.Close()

		var tasks []models.Task
		for rows.Next() {
			var task models.Task
			err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
			if err != nil {
				fmt.Println(op, "Ошибка сканирования строки:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("%s: Ошибка обработки данных: %v", op, err)})
			}
			tasks = append(tasks, task)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": tasks,
		})
	}
}

func GetTaskById(db *pgx.Conn) fiber.Handler {
	const op = "handlers.task.GetTaskById"
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := `SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1`

		var task models.Task

		err := db.QueryRow(context.Background(), query, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("%s  Failed to get task into the database %v", op, err)})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": task,
		})
	}
}

func CreateTask(db *pgx.Conn) fiber.Handler {
	const op = "handlers.task.CreateTask"
	return func(c *fiber.Ctx) error {
		var task models.Task
		if err := c.BodyParser(&task); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s: Failed to parse request body: %v", op, err),
			})
		}

		if task.Title == "" || task.Description == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Title and description are required fields",
			})
		}

		task.Status = "new" // or another default status
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()

		query := `
			INSERT INTO tasks (title, description, status, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id, created_at, updated_at`

		var createdTask models.Task
		err := db.QueryRow(context.Background(), query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt).
			Scan(&createdTask.ID, &createdTask.CreatedAt, &createdTask.UpdatedAt)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("%s: Failed to insert task into the database: %v", op, err),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"data": createdTask,
		})
	}
}

func DeleteTask(db *pgx.Conn) fiber.Handler {
	const op = "handlers.task.DeleteTask"
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := `DELETE FROM tasks WHERE id = $1`
		_, err := db.Exec(context.Background(), query, id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("%s: Failed to delete task from the database: %v", op, err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Task deleted successfully",
		})
	}
}

func UpdateTask(db *pgx.Conn) fiber.Handler {
	const op = "handlers.task.UpdateTask"
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var task models.Task
		if err := c.BodyParser(&task); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s: Failed to parse request body: %v", op, err),
			})
		}

		if task.Title == "" || task.Description == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Title and description are required fields"),
			})
		}

		query := `UPDATE tasks
			SET title = $1, description = $2, status = $3, updated_at = $4
			WHERE id = $5
			RETURNING id, title, description, status, updated_at`

		var updatedTask models.Task
		err := db.QueryRow(context.Background(), query, task.Title, task.Description, task.Status, id).Scan(
			&updatedTask.ID,
			&updatedTask.Title,
			&updatedTask.Description,
			&updatedTask.Status,
			&updatedTask.UpdatedAt,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("%s: Failed to update task into the database: %v", op, err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Task update successfully",
			"data":    updatedTask,
		})
	}
}
