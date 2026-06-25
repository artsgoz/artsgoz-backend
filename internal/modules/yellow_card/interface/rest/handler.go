package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type YellowCardHandler struct {
	getYellowCard  *application.GetYellowCard
	updateProfile  *application.UpdateProfile
	updateSubjects *application.UpdateSubjects
	recordPDPA     *application.RecordPDPA
}

func NewYellowCardHandler(
	getYellowCard *application.GetYellowCard,
	updateProfile *application.UpdateProfile,
	updateSubjects *application.UpdateSubjects,
	recordPDPA *application.RecordPDPA,
) *YellowCardHandler {
	return &YellowCardHandler{
		getYellowCard:  getYellowCard,
		updateProfile:  updateProfile,
		updateSubjects: updateSubjects,
		recordPDPA:     recordPDPA,
	}
}

func (h *YellowCardHandler) Get(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}

	card, err := h.getYellowCard.Execute(c.Context(), uid)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    card,
	})
}

func (h *YellowCardHandler) UpdateProfile(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}

	var in application.UpdateProfileInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	if err := h.updateProfile.Execute(c.Context(), uid, in); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Student profile updated successfully",
	})
}

func (h *YellowCardHandler) UpdateSubjects(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}

	var in application.UpdateSubjectsInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	if err := h.updateSubjects.Execute(c.Context(), uid, in); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Yellow card subjects and GPA data updated successfully",
	})
}

func (h *YellowCardHandler) RecordPDPA(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}

	var in application.RecordPDPAInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	if err := h.recordPDPA.Execute(c.Context(), uid, in.Agreed); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "PDPA consent recorded successfully",
	})
}
