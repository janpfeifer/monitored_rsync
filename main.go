package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/janpfeifer/must"
	klog "k8s.io/klog/v2"
)

var (
	flagExclude = flag.String("exclude", "", "List (comma-separated) of subdirectories (from the "+
		"source directory) to exclude from monitoring and rsync.")
	flagDelay = flag.Int("delay", 1000, "Time in milliseconds to wait for changes, before invoking rsync. "+
		"This is usually efficient because changes to files usually happen in burst, and we want to avoid multiple "+
		"rsync invocations.")
)

func AssertNoError(err error) {
	if err != nil {
		log.Fatalf("Failed: %+v", err)
	}
}

// ReplaceTildeInDir by the user's home directory. Returns dir if it doesn't start with "~".
func ReplaceTildeInDir(dir string) string {
	if len(dir) == 0 || dir[0] != '~' {
		return dir
	}
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	return path.Join(homeDir, dir[1:])
}

func AbsoluteSourceDirectory(srcDir string) string {
	srcDir = ReplaceTildeInDir(srcDir)
	if srcDir[0] != '/' {
		if strings.HasPrefix(srcDir, "./") {
			srcDir = srcDir[2:]
		}
		pwd := must.M1(os.Getwd())
		if srcDir == "." {
			srcDir = pwd
		} else {
			srcDir = path.Join(pwd, srcDir)
		}
	}
	return srcDir
}

var (
	style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(4).
		PaddingRight(4)
)

func main() {
	klog.InitFlags(nil)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, `$ monitored_rsync [flags...] <source_directory> <remote_target>

  It monitors changes in directories using inotifywait, and when changes happen, it invokes rsync. 

  <source_directory>: where to monitor and rsync from.
  <remove_target>: passed to rsync

  Use '--vmodule=rsync=2' to see the rsync command executed. And use '--vmodule=monitor=2' to see file change
  events received.

`)
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		klog.Exitf("%s requires 2 arguments!", os.Args[0])
	}
	sourceDirectory := AbsoluteSourceDirectory(args[0])
	remoteDirectory := args[1]
	var excludePaths []string
	if *flagExclude != "" {
		excludePaths = strings.Split(*flagExclude, ",")
		excludePaths = slices.DeleteFunc(excludePaths, func(s string) bool { return s == "" })
	}

	// Verbose.
	headerStyle := lipgloss.NewStyle().
		PaddingRight(3).PaddingLeft(3).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("99")).
		Bold(true)
	parts := []string{
		fmt.Sprintf("Source Directory: \t%s", sourceDirectory),
		fmt.Sprintf("Remote Directory: \t%s", remoteDirectory),
	}
	if len(excludePaths) > 0 {
		parts = append(parts, fmt.Sprintf("Exclude paths:"))
		for _, p := range excludePaths {
			parts = append(parts, fmt.Sprintf("\t%s", p))
		}
	}
	fmt.Println(headerStyle.Render(strings.Join(parts, "\n")))

	err := Monitor(sourceDirectory, excludePaths, time.Millisecond*time.Duration(*flagDelay), func() error {
		fmt.Println(style.Render(fmt.Sprintf("rsync @ %s", time.Now())))
		return RSync(sourceDirectory, remoteDirectory, excludePaths)
	})
	if err != nil {
		klog.Exitf("Failed: %+v", err)
	}
	return
}
