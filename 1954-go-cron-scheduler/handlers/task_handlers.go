package handlers

import (
	"net/http"

	"cron-scheduler/models"
	"cron-scheduler/scheduler"

	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	scheduler *scheduler.Scheduler
}

func NewTaskHandler(s *scheduler.Scheduler) *TaskHandler {
	return &TaskHandler{scheduler: s}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var req models.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	if req.CronExpr == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "cron_expr is required",
		})
	}

	if req.Type != models.TaskTypeCommand && req.Type != models.TaskTypeHTTP {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "type must be 'command' or 'http'",
		})
	}

	if req.Type == models.TaskTypeCommand && req.Command == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "command is required for command type tasks",
		})
	}

	if req.Type == models.TaskTypeHTTP && req.CallbackURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "callback_url is required for http type tasks",
		})
	}

	task := models.NewTask(&req)

	if err := h.scheduler.AddTask(task); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(task.ToResponse())
}

func (h *TaskHandler) ListTasks(c *fiber.Ctx) error {
	tasks := h.scheduler.ListTasks()
	responses := make([]*models.TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = task.ToResponse()
	}
	return c.JSON(responses)
}

func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	id := c.Params("id")
	task, exists := h.scheduler.GetTask(id)
	if !exists {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}
	return c.JSON(task.ToDetailResponse())
}

func (h *TaskHandler) PauseTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.scheduler.PauseTask(id); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	task, _ := h.scheduler.GetTask(id)
	return c.JSON(task.ToResponse())
}

func (h *TaskHandler) ResumeTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.scheduler.ResumeTask(id); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	task, _ := h.scheduler.GetTask(id)
	return c.JSON(task.ToResponse())
}

func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.scheduler.DeleteTask(id); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "task deleted",
	})
}

func (h *TaskHandler) GetTaskLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	logs, err := h.scheduler.GetTaskLogs(id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(logs)
}
