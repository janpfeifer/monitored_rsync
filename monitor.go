package main

import (
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
	"os"
	"path"
	"path/filepath"
	"time"
)

// RecursiveWatcher creates a new fsnotify Watcher for the specified path and its subdirectories.
func RecursiveWatcher(watchPath string, excludedPaths []string) (*fsnotify.Watcher, error) {
	// Create excluded set from list.
	excludedPathsSet := make(map[string]struct{})
	for _, exclude := range excludedPaths {
		if exclude[0] != '/' {
			exclude = path.Join(watchPath, exclude)
		}
		excludedPathsSet[exclude] = struct{}{}
	}

	// Create watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Recursively travel tree, and collect directories to watch.
	err = filepath.Walk(watchPath, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded paths
		if _, ok := excludedPathsSet[newPath]; ok {
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
		AssertNoError(watcher.Close())
		return nil, err
	}

	return watcher, nil
}

func Monitor(srcDir string, excludePaths []string, delay time.Duration, onChange func() error) error {
	watcher := MustNoError(RecursiveWatcher(srcDir, excludePaths))
	var err error
	for {
		// Wait for anything to change.
		select {
		case ev := <-watcher.Events:
			klog.V(2).Infof("watcher: %s", ev)
			// Something changed, start timer.
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
					// Nothing to do, timer already running.
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

//
//func main() {
//	excludedPaths := make(map[string]struct{})
//	excludedPaths["./exclude-dir"] = struct{}{}
//
//	watcher, err := RecursiveWatcher(".", excludedPaths)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	defer watcher.Close()
//
//	done := make(chan bool)
//	go func() {
//		for {
//			select {
//			case event, ok := <-watcher.Events:
//				if !ok {
//					return
//				}
//				log.Println("event:", event)
//				if event.Op&fsnotify.Write == fsnotify.Write {
//					log.Println("modified file:", event.Name)
//				}
//			case err, ok := <-watcher.Errors:
//				if !ok {
//					return
//				}
//				log.Println("error:", err)
//			}
//		}
//	}()
//
//	<-done
//}
