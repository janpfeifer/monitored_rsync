package main

import (
	"flag"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
)

var (
	flagDryRun = flag.Bool("dry-run", false, "If set to true, it will run rsync in dry-run mode.")
)

func RSync(srcDir, remote string, excludePaths []string) error {
	klog.V(2).Infof("RSync(%s, %s, %v)", srcDir, remote, excludePaths)
	args := []string{
		"--archive",
		"--delete",
		"--human-readable",
		"--verbose",
		"--update",
	}
	if *flagDryRun {
		args = append(args, "--dry-run")
	}
	for _, exclude := range excludePaths {
		args = append(args, "--exclude", exclude)
	}
	args = append(args, ".", remote)
	cmd := exec.Command("rsync", args...)
	cmd.Dir = srcDir
	klog.V(2).Infof("rsync command: %s", cmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
