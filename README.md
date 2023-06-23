# monitored_rsync

This simple command line tool simply call a `inotifywait` on a directory (recursively
tracking subdirectories) and whenever there is a change, it calls `rsync`.

There are many uses for this, including a live backup system, a synchronizing way
to develop on a local machine (let's say a laptop), and have the code sync'ed up
to a server, where presumably compilation happens. Notice this is often better
than something like `sshfs` because it has a copy of all files locally, which
greatly speeds up IDEs (`sshfs` over high-latency connections is annoying to
develop, specially for IDEs that are constantly checking if files changed).

## Installation

```bash
$ go install github.com/janpfeifer/monitored_rsync@latest
```

Note: prefix with `GOPROXY=direct` if you want to make sure it is getting the latest version.

## Example:

```bash
$ monitored_rsync --exclude=.git,.idea ~/Projects/MyProject me@myhost:Projects/MyProject
```

## More options

From --help:

```text
Usage of monitored_rsync:
$ monitored_rsync [flags...] <source_directory> <remote_target>
                                           
  It monitors changes in directories using inotifywait, and when changes happen, it invokes rsync.                                              
                                           
  <source_directory>: where to monitor and rsync from.    
  <remove_target>: passed to rsync
                                                                                                                                                                              
  Use '--vmodule=rsync=2' to see the rsync command executed. And use '--vmodule=monitor=2' to see file change
  events received.                                                                                                                                                            
                                           
  -delay int
        Time in milliseconds to wait for changes, before invoking rsync. This is usually efficient because changes to files usually happen in burst, and we want to avoid mult
iple rsync invocations. (default 1000)
  -dry-run
        If set to true, it will run rsync in dry-run mode.
  -exclude string
        List (comma-separated) of subdirectories (from the source directory) to exclude from monitoring and rsync.
```
_(Various logging flags omitted)_

## TODO

The `rsync` flags are hard-coded for now :( -- except the `--exclude` and `--dry-run`. 
Someone should create a flag to make it configurable.

