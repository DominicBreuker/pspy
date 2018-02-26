package inotify

import "golang.org/x/sys/unix"

var InotifyEvents = map[uint32]string{
	unix.IN_ACCESS:                          "ACCESS",
	unix.IN_ATTRIB:                          "ATTRIB",
	unix.IN_CLOSE_NOWRITE:                   "CLOSE_NOWRITE",
	unix.IN_CLOSE_WRITE:                     "CLOSE_WRITE",
	unix.IN_CREATE:                          "CREATE",
	unix.IN_DELETE:                          "DELETE",
	unix.IN_DELETE_SELF:                     "DELETE_SELF",
	unix.IN_MODIFY:                          "MODIFY",
	unix.IN_MOVED_FROM:                      "MOVED_FROM",
	unix.IN_MOVED_TO:                        "MOVED_TO",
	unix.IN_MOVE_SELF:                       "MOVE_SELF",
	unix.IN_OPEN:                            "OPEN",
	(unix.IN_ACCESS | unix.IN_ISDIR):        "ACCESS DIR",
	(unix.IN_ATTRIB | unix.IN_ISDIR):        "ATTRIB DIR",
	(unix.IN_CLOSE_NOWRITE | unix.IN_ISDIR): "CLOSE_NOWRITE DIR",
	(unix.IN_CLOSE_WRITE | unix.IN_ISDIR):   "CLOSE_WRITE DIR",
	(unix.IN_CREATE | unix.IN_ISDIR):        "CREATE DIR",
	(unix.IN_DELETE | unix.IN_ISDIR):        "DELETE DIR",
	(unix.IN_DELETE_SELF | unix.IN_ISDIR):   "DELETE_SELF DIR",
	(unix.IN_MODIFY | unix.IN_ISDIR):        "MODIFY DIR",
	(unix.IN_MOVED_FROM | unix.IN_ISDIR):    "MOVED_FROM DIR",
	(unix.IN_MOVE_SELF | unix.IN_ISDIR):     "MODE_SELF DIR",
	(unix.IN_OPEN | unix.IN_ISDIR):          "OPEN DIR",
}
