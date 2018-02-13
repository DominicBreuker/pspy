package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dominicbreuker/pspy/internal/pspy"
	"github.com/spf13/cobra"
)

var bannerLines = []string{
	"      _____   _____ _______     __",
	"     |  __ \\ / ____|  __ \\ \\   / /",
	"     | |__) | (___ | |__) \\ \\_/ / ",
	"     |  ___/ \\___ \\|  ___/ \\   /  ",
	"     | |     ____) | |      | |   ",
	"     |_|    |_____/|_|      |_|   ",
	helpText,
}

var helpText = `
pspy monitors the system for file system events and new processes.
It prints out these envents to the console.
File system events are monitored with inotify.
Processes are monitored by scanning /proc, using file system events as triggers.
pspy does not require root permissions do operate.
Check our https://github.com/dominicbreuker/pspy for more information.
`

var banner = strings.Join(bannerLines, "\n")

var rootCmd = &cobra.Command{
	Use:   "pspy",
	Short: "pspy can watch your system for new processes and file system events",
	Long:  banner,
	Run:   root,
}

var logPS, logFS bool
var rDirs, dirs []string

func init() {
	rootCmd.PersistentFlags().BoolVarP(&logPS, "procevents", "p", true, "print new processes to stdout")
	rootCmd.PersistentFlags().BoolVarP(&logFS, "fsevents", "f", true, "print file system events to stdout")
	rootCmd.PersistentFlags().StringArrayVarP(&rDirs, "recursive_dirs", "r", []string{"/tmp"}, "watch these dirs recursively")
	rootCmd.PersistentFlags().StringArrayVarP(&dirs, "dirs", "d", []string{}, "watch these dirs")
}

func root(cmd *cobra.Command, args []string) {
	fmt.Printf("Watching recursively    : %+v (%d)\n", rDirs, len(rDirs))
	fmt.Printf("Watching non-recursively: %+v (%d)\n", dirs, len(dirs))
	fmt.Printf("Printing: processes=%t file-system events:%t\n", logPS, logFS)
	pspy.Watch(rDirs, dirs, logPS, logFS)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
