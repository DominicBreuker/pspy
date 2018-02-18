<img src="images/logo.svg" align="left" />

# pspy - unprivileged linux process snooping

pspy is a command line tool designed to snoop on processes without needing root permissions.
It allows you to see commands run by other users, cron jobs, etc. as they execute.
Great for enumeration of linux systems in CTFs.

The tool gathers it's info from procfs scans.
Inotify watchers placed on selected parts of the file system trigger these scans to increase the chance of catching short-lived processes.

## Getting started

Get the tool onto the Linux machine you want to inspect.
You must choose between the 32 and 64 bit versions.
The files are (for now) in the `/bin` folder of this repository.
- 32 bit version: [download](bin/pspy32)
- 64 bit version: [download](bin/pspy64)

You can run `pspy --help` to learn about the flags and their meaning.
The summary is as follows:
- -p: enables printing commands to stdout (enabled by default)
- -f: enables printing file system events to stdout (disabled by default)
- -r: list of directories to watch with inotify. pspy will watch all subdirectories recursively (by default, watches /usr, /tmp, /etc, /home, /var, and /opt).
- -d: list of directories to watch with inotify. pspy will watch these directories only, not the subdirectories (empty by default).

Default settings should be fine for most applications.
Watching files inside `/usr` is most important since many tools will access libraries inside it.

Some more complex examples:

```bash
# print both commands and file system events, but watch only two directories (one recursive, one not)
pspy64 -pf -r /path/to/my/dir -d /path/to/my/other/dir

# disable printing commands but enable file system events
pspy64 -p=false -f
```

### Examples

### Cron job watching

To see the tool in action, just clone the repo and run `make example` (Docker needed).
The example starts a debian container in which a cron job changes a user password every minute.
After starting cron, it runs pspy in foreground.
You should see output similar to this:

```console
~/pspy (master) $ make example
[...]
docker run -it --rm local/pspy-example:latest
[+] cron started
[+] Running as user uid=1000(myuser) gid=1000(myuser) groups=1000(myuser),27(sudo)
[+] Starting pspy now...
Watching recursively    : [/usr /tmp /etc /home /var /opt] (6)
Watching non-recursively: [] (0)
Printing: processes=true file-system events=false
2018/02/18 21:00:03 Inotify watcher limit: 524288 (/proc/sys/fs/inotify/max_user_watches)
2018/02/18 21:00:03 Inotify watchers set up: Watching 1030 directories - watching now
2018/02/18 21:00:03 CMD: UID=0    PID=9      | cron -f
2018/02/18 21:00:03 CMD: UID=0    PID=7      | sudo cron -f
2018/02/18 21:00:03 CMD: UID=1000 PID=14     | pspy
2018/02/18 21:00:03 CMD: UID=1000 PID=1      | /bin/bash /entrypoint.sh
2018/02/18 21:01:01 CMD: UID=0    PID=20     | CRON -f
2018/02/18 21:01:01 CMD: UID=0    PID=21     | CRON -f
2018/02/18 21:01:01 CMD: UID=0    PID=22     | python3 /root/scripts/password_reset.py
2018/02/18 21:01:01 CMD: UID=0    PID=25     |
2018/02/18 21:01:01 CMD: UID=???  PID=24     | ???
2018/02/18 21:01:01 CMD: UID=0    PID=23     | /bin/sh -c /bin/echo -e "KI5PZQ2ZPWQXJKEL\nKI5PZQ2ZPWQXJKEL" | passwd myuser
2018/02/18 21:01:01 CMD: UID=0    PID=26     | /usr/sbin/sendmail -i -FCronDaemon -B8BITMIME -oem root
2018/02/18 21:01:01 CMD: UID=101  PID=27     |
2018/02/18 21:01:01 CMD: UID=8    PID=28     | /usr/sbin/exim4 -Mc 1enW4z-00000Q-Mk
```

First, pspy prints all currently running processes.
It prints PID, UID and the command line.
Each time pspy detects a new PID, it adds a line to this log.
In this example, you find a process with PID 23 which seems to change the password of myuser.
This is the result of a Python script used in roots private crontab `/var/spool/cron/crontabs/root`, which executes this shell command (check [crontab](docker/var/spool/cron/crontabs/root) and [script](docker/root/scripts/password_reset.py)).
Note that myuser can neither see the crontab nor the Python script.
With pspy, it can see the commands nevertheless.

### CTF example from Hach The Box

Below is an example from the machine Shrek from [Hack The Box](https://www.hackthebox.eu/).
In this CTF challenge, the task is to exploit a hidden cron job that's changing ownership of all files in a folder.
With pspy, the cron job is easy to find and analyse:

![animated demo gif](images/demo.gif)

## How it works

Several tools exist to list all processes executed on Linux systems, including those that have finished.
For instance there is [forkstat](http://smackerelofopinion.blogspot.de/2014/03/forkstat-new-tool-to-trace-process.html).
It receives notifications from the kernel on process-related events such as fork and exec.

Unfortunately, the tool requires root privileges so you cannot use it to right away.
However, nothing stop you in general from snooping on the processes running on the system.
All data is visible as long as the process is running.
The only problem is you have to catch short-lived processes in the very short time span in which they are alive.
Scanning the `/proc` directory for new PIDs in an infinite loop does the trick but consumes a lot of CPU.

A stealthier way is to use the following trick.
Process tend to access files such as libraries in `/usr`, temporary files in `/tmp`, log files in `/var`, ...
Without root permissions, you can get notifications whenever these files are touched.
The API for this is called [inotify](http://man7.org/linux/man-pages/man7/inotify.7.html).
While we cannot monitor processes directly, but we can monitor their interactions with the file system.

We can use the file system events as a trigger to scan `/proc`, hoping that we can do it fast enough to catch the processes.
This is what pspy does.
Thus, there is no guarantee you won't miss one, but chances seem to be good in my experiments.
In general, the longer the processes run, the bigger the chance of catching them is.

Besides using the events, pspy will also scan `/proc` every 100ms, just to be sure.
CPU usage seems to be quite low for this interval.
Making the interval configurable is on the roadmap.

# Misc

Logo: "By Creative Tail [CC BY 4.0 (http://creativecommons.org/licenses/by/4.0)], via Wikimedia Commons" ([link](https://commons.wikimedia.org/wiki/File%3ACreative-Tail-People-spy.svg))
