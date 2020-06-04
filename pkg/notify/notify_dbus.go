// +build linux

package notify

import (
	"context"
	"os"

	"github.com/godbus/dbus"

	"passline/pkg/ctxutil"
)

// Notify displays a desktop notification with dbus
func Notify(ctx context.Context, subj, msg string) error {
	if os.Getenv("PASSLINE_NO_NOTIFY") != "" || !ctxutil.IsNotifications(ctx) {
		// debug.Log("Notifications disabled")
		return nil
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		// debug.Log("DBus failure: %s", err)
		return err
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, "passline", uint32(0), iconURI(), subj, msg, []string{}, map[string]dbus.Variant{}, int32(5000))
	if call.Err != nil {
		// debug.Log("DBus notification failure: %s", call.Err)
		return err
	}

	return nil
}
