package datastar

import (
	"github.com/gofiber/fiber/v2"
)

// ModalConfig holds the configuration for opening a modal.
type ModalConfig struct {
	Title       string `json:"title"`
	ShowFooter  bool   `json:"showFooter"`
	ShowConfirm bool   `json:"showConfirm"`
	ConfirmText string `json:"confirmText,omitempty"`
}

// modalSignal is the shape Datastar pushes to the client store.
type modalSignal struct {
	Modal struct {
		Visible     bool   `json:"visible"`
		ShowFooter  bool   `json:"showFooter"`
		ShowConfirm bool   `json:"showConfirm"`
		ConfirmText string `json:"confirmText,omitempty"`
	} `json:"modal"`
}

// OpenModal opens a modal with the specified configuration.
//
//	config := datastar.ModalConfig{
//	    Title:       "Confirm Action",
//	    ShowFooter:  true,
//	    ShowConfirm: true,
//	    ConfirmText: "Delete",
//	}
//	datastar.OpenModal(c, config)
//	datastar.MergeFragmentTempl(c, views.ConfirmContent(item))
func OpenModal(c *fiber.Ctx, config ModalConfig) error {
	var s modalSignal
	s.Modal.Visible = true
	s.Modal.ShowFooter = config.ShowFooter
	s.Modal.ShowConfirm = config.ShowConfirm
	s.Modal.ConfirmText = config.ConfirmText
	return MarshalAndMergeSignals(c, s)
}

// CloseModal closes the currently open modal.
//
//	datastar.CloseModal(c)
func CloseModal(c *fiber.Ctx) error {
	var s modalSignal
	s.Modal.Visible = false
	return MarshalAndMergeSignals(c, s)
}

// ToggleModal toggles the modal visibility state.
//
//	datastar.ToggleModal(c)
func ToggleModal(c *fiber.Ctx) error {
	return MarshalAndMergeSignals(c, map[string]any{
		"modal": map[string]any{
			"visible": "!$modal.visible",
		},
	})
}
