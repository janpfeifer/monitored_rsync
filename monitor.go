package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/janpfeifer/must"
	"k8s.io/klog/v2"
	"os"
	"path"
	"path/filepath"
	"time"
)

var empty = struct{}{}

// RecursiveWatcher creates a new fsnotify Watcher for the specified path and its subdirectories.
func RecursiveWatcher(watchPath string, excludedPaths []string) (*fsnotify.Watcher, Set[string], error) {
	// Create excluded set from list.
	excludedPathsSet := MakeSet[string](len(excludedPaths))
	for _, exclude := range excludedPaths {
		if !path.IsAbs(exclude) {
			exclude = path.Join(watchPath, exclude)
		}
		excludedPathsSet.Insert(exclude)
	}

	// Create watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	// Recursively travel tree, and collect directories to watch.
	err = filepath.Walk(watchPath, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded paths
		if excludedPathsSet.Has(newPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			err = watcher.Add(newPath)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		must.M(watcher.Close())
		return nil, nil, err
	}

	return watcher, excludedPathsSet, nil
}

// UpdateWatcher updates the watcher to new directories that may have been created.
func UpdateWatcher(watcher *fsnotify.Watcher, excludeSet Set[string], ev fsnotify.Event) error {
	if !ev.Has(fsnotify.Create) {
		return nil
	}
	if excludeSet.Has(ev.Name) {
		return nil
	}
	fi, err := os.Stat(ev.Name)
	if err != nil {
		// Non-fatal error.
		klog.Errorf("Failed to Stat(%q): %+v", ev.Name, err)
		return nil
	}
	if !fi.IsDir() {
		return nil
	}
	err = watcher.Add(ev.Name)
	if err != nil {
		// Non-fatal error.
		klog.Errorf("Failed to watcher.Add(%q): %+v", ev.Name, err)
		return nil
	}
	return nil
}

// Monitor watches for change events on the sub-directories under srcDir, and call onChange if anything changes.
//
// Limitations: it doesn't monitor new directories that may be created.
//
// The excludePaths are not monitored.
func Monitor(srcDir string, excludePaths []string, delay time.Duration, onChange func() error) error {
	watcher, excludeSet := must.M2(RecursiveWatcher(srcDir, excludePaths))
	var err error
	for {
		// Wait for anything to change.
		select {
		case ev := <-watcher.Events:
			klog.V(2).Infof("watcher: %s", ev)
			// Something changed, start timer.
			err = UpdateWatcher(watcher, excludeSet, ev)
			if err != nil {
				return err
			}
		case err = <-watcher.Errors:
			return err
		}

		// Wait for delay before calling onChange.
		if delay > 0 {
			timer := time.NewTimer(delay)
		waitOnChange:
			for {
				select {
				case ev := <-watcher.Events:
					klog.V(2).Infof("watcher (on timer): %s", ev)
					err = UpdateWatcher(watcher, excludeSet, ev)
					if err != nil {
						return err
					}
					timer = time.NewTimer(delay) // Reset timer for more changes.

				case err = <-watcher.Errors:
					return err

				case <-timer.C:
					klog.V(2).Info("watcher: calling onChange")
					break waitOnChange
				}
			}
		}

		// Now call onChange.
		err = onChange()
		if err != nil {
			return err
		}
	}
}
