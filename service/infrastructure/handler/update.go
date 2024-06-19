package handler

import (
	"encoding/base64"
	"log"

	d "github.com/bruno5200/CSM/service/domain"
	p "github.com/bruno5200/CSM/service/infrastructure/presenter"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *serviceHandler) Put(c *fiber.Ctx) error {

	id, err := uuid.Parse(c.Params("serviceId"))

	if err != nil {
		log.Printf("Error parsing id: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(p.ServiceErrorResponse(d.ErrInvalidServiceId))
	}

	serviceRequest, err := d.UnmarshalServiceRequest(c.Body())

	if err != nil {
		log.Printf("Error unmarshalling service request: %s", err)
		return c.Status(fiber.StatusBadRequest).JSON(p.ServiceErrorResponse(err))
	}

	service, err := h.serviceService.GetService(id)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(p.ServiceErrorResponse(err))
	}

	service.Update(serviceRequest)

	if err := h.serviceService.UpdateService(service); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(p.ServiceErrorResponse(err))
	}

	service.Key = base64.RawStdEncoding.EncodeToString([]byte(service.Key))

	return c.Status(fiber.StatusOK).JSON(p.ServiceSuccessResponse(service))
}