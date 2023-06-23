package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

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

func MustNoError[T any](value T, err error) T {
	AssertNoError(err)
	return value
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
		pwd := MustNoError(os.Getwd())
		if srcDir == "." {
			srcDir = pwd
		} else {
			srcDir = path.Join(pwd, srcDir)
		}
	}
	return srcDir
}

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
	}

	// Verbose.
	fmt.Printf("Source directory:\t%s\n", sourceDirectory)
	fmt.Printf("Remote directory:\t%s\n", remoteDirectory)
	if len(excludePaths) > 0 {
		fmt.Printf("Exclude paths:   \t%v\n", excludePaths)
	}

	err := Monitor(sourceDirectory, excludePaths, time.Millisecond*time.Duration(*flagDelay), func() error {
		return RSync(sourceDirectory, remoteDirectory, excludePaths)
	})
	if err != nil {
		klog.Exitf("Failed: %+v", err)
	}
	return
}
