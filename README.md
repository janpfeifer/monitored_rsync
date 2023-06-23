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

Example:

```bash
$ monitored_rsync --exclude=.git,.idea ~/Projects/MyProject me@myhost:Projects/MyProject
```

## More options

See --help for several options.

## TODO

RSync flags are hard-coded for now :( -- except the `--exclude` and `--dry-run`. 
Someone should create a flag to make it configurable.

