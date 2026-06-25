package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/application"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type CreditTrackingHandler struct {
	getSubjects    *application.GetSubjects
	updateSubjects *application.UpdateSubjectCompletion
	addSubject     *application.AddCustomSubject
	deleteSubject  *application.DeleteCustomSubject
	getProfile     *application.GetProfile
	updateProfile  *application.UpdateProfile
}

func NewCreditTrackingHandler(
	getSubjects *application.GetSubjects,
	updateSubjects *application.UpdateSubjectCompletion,
	addSubject *application.AddCustomSubject,
	deleteSubject *application.DeleteCustomSubject,
	getProfile *application.GetProfile,
	updateProfile *application.UpdateProfile,
) *CreditTrackingHandler {
	return &CreditTrackingHandler{
		getSubjects:    getSubjects,
		updateSubjects: updateSubjects,
		addSubject:     addSubject,
		deleteSubject:  deleteSubject,
		getProfile:     getProfile,
		updateProfile:  updateProfile,
	}
}

func (h *CreditTrackingHandler) GetSubjects(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}
	out, err := h.getSubjects.Execute(c.Context(), uid)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *CreditTrackingHandler) UpdateSubjectCompletion(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}
	subjectID := c.Params("subjectId")

	var in application.ToggleCompletionInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.updateSubjects.Execute(c.Context(), uid, subjectID, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subject completion status updated successfully",
		"data":    out,
	})
}

func (h *CreditTrackingHandler) AddCustomSubject(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}

	var in application.CustomSubjectInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.addSubject.Execute(c.Context(), uid, in)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Custom subject added successfully",
		"data":    out,
	})
}

func (h *CreditTrackingHandler) DeleteCustomSubject(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}
	subjectID := c.Params("subjectId")

	if err := h.deleteSubject.Execute(c.Context(), uid, subjectID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Custom subject deleted successfully",
	})
}

func (h *CreditTrackingHandler) GetProfile(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}
	out, err := h.getProfile.Execute(c.Context(), uid)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

func (h *CreditTrackingHandler) UpdateProfile(c fiber.Ctx) error {
	uid, ok := c.Locals("uid").(string)
	if !ok || uid == "" {
		return apperr.Unauthorized("unauthorized: missing user context")
	}

	var in application.ProfileInput
	if err := c.Bind().Body(&in); err != nil {
		return apperr.Validation("invalid request body", err.Error())
	}

	out, err := h.updateProfile.Execute(c.Context(), uid, in)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Academic profile updated successfully",
		"data":    out,
	})
}
