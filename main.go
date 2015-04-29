package main

import (
        "errors"
        "flag"
        "fmt"
        "log"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "time"

        "github.com/howeyc/fsnotify"
)

var (
        updatedAt time.Time
        interval  time.Duration
        patterns  []string

        // cli args
        cmd string

        // cli options
        dir    string
        n      float64
        hidden bool
        p      string
)

func loop() {
        watcher, err := fsnotify.NewWatcher()
        if err != nil {
                log.Fatal(err)
        }
        err = watcher.Watch(dir)
        if err != nil {
                log.Fatal(err)
        }

        updatedAt = time.Now()
        for {
                select {
                case ev := <-watcher.Event:
                        if updatedAt.Add(interval).After(time.Now()) {
                                break
                        }
                        if handle(ev) != nil {
                                break
                        }
                        updatedAt = time.Now()
                case err := <-watcher.Error:
                        log.Println("ift error:", err)
                }
        }
}

func handle(ev *fsnotify.FileEvent) error {
        if err := watched(ev.Name); err != nil {
                return err
        }
        c := exec.Command("sh", "-c", cmd)
        c.Dir = dir
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr
        name, _ := filepath.Rel(dir, ev.Name)
        log.Println(name+":", cmd)
        if err := c.Run(); err != nil {
                log.Println(err)
        }
        return nil
}

func watched(path string) error {
        if path == "" {
                return errors.New("file name not found")
        }
        if !hidden && filepath.Base(path)[0] == '.' {
                return errors.New("hidden file")
        }
        // watch all
        if len(patterns) == 1 && patterns[0] == "" {
                return nil
        }
        var err error
        path, err = filepath.Rel(dir, path)
        if err != nil {
                return err
        }

        for _, p := range patterns {
                m, err := filepath.Match(p, path)
                if err != nil || !m {
                        continue
                }
                return nil
        }
        return errors.New("file is not being watched")
}

func loadWatchFile() {
        // pass
}

func main() {
        flag.Usage = func() {
                fmt.Println("ift [-d dir] [-n secs] [-p patterns] [-hidden] command")
                fmt.Println("\nOPTIONS:")
                flag.PrintDefaults()
        }
        flag.StringVar(&dir, "d", ".", "Watch directory")
        flag.Float64Var(&n, "n", 1.0, "Interval seconds")
        flag.BoolVar(&hidden, "hidden", false, "Watch hidden file")
        flag.StringVar(&p, "p", "", "Specify file name patterns to watch. "+
                "Multiple patterns should be seperated by comma. "+
                "If pattern is not specified all files in the dir will be watched")
        flag.Parse()

        dir, _ = filepath.Abs(dir)
        interval = time.Duration(n) * time.Second
        patterns = strings.Split(p, ",")

        if flag.NArg() != 1 {
                fmt.Println("Command not found")
                os.Exit(1)
        }

        cmd = flag.Arg(0)
        loop()
}
