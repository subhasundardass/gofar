// framework/datastar/toast.go
package datastar

import "github.com/gofiber/fiber/v2"

// ToastLevel controls the visual style of the notification.
type ToastLevel string

const (
	ToastSuccess ToastLevel = "success"
	ToastError   ToastLevel = "error"
	ToastWarning ToastLevel = "warning"
	ToastInfo    ToastLevel = "info"
)

// toastSignal is the shape Datastar pushes to the client store.
type toastSignal struct {
	Toast struct {
		Message string     `json:"message"`
		Level   ToastLevel `json:"level"`
		Visible bool       `json:"visible"`
	} `json:"toast"`
}

// Toast pushes a toast notification signal to the client.
// The frontend reacts to $toast.visible becoming true and renders the message.
//
//	datastar.Toast(c, datastar.ToastSuccess, "Country saved")
//	datastar.Toast(c, datastar.ToastError, "Something went wrong")
func Toast(c *fiber.Ctx, level ToastLevel, message string) error {
	var s toastSignal
	s.Toast.Message = message
	s.Toast.Level = level
	s.Toast.Visible = true
	return MarshalAndMergeSignals(c, s)
}
