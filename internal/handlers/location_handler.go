package handlers

import (
	"strconv"
	"transjakarta-fleet-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type locationHandler struct {
	Repo *repository.LocationRepository
}

func NewLocationHandler(repo *repository.LocationRepository) *locationHandler {
	return &locationHandler{Repo: repo}
}

func (h *locationHandler) GetLatestLocation(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")

	loc, err := h.Repo.GetLatestLocation(vehicleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Location not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(loc)
}

func (h *locationHandler) GetLocationHistory(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")

	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start timestamp",
		})
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end timestamp",
		})
	}

	loc, err := h.Repo.GetLocationHistory(vehicleID, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(loc)
}