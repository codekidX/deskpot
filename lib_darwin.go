package deskpot

import (
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

type NSUserNotification struct {
	objc.Object
}

var NSUserNotification_ = objc.Get("NSUserNotification")

type NSUserNotificationCenter struct {
	objc.Object
}

var NSUserNotificationCenter_ = objc.Get("NSUserNotificationCenter")

type Notification struct {
	Title       string
	Information string
	Icon        string
}

func Notify(n Notification) {
	noti := NSUserNotification{NSUserNotification_.Alloc().Init()}
	noti.Set("title:", core.String(n.Title))
	noti.Set("informativeText:", core.String(n.Information))

	center := NSUserNotificationCenter{NSUserNotificationCenter_.Send("defaultUserNotificationCenter")}
	center.Send("deliverNotification:", noti)
	noti.Release()
}
