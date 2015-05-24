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

        "github.com/go-fsnotify/fsnotify"
)

var (
        patterns = []string(nil)
        cmd      string

        // cli options
        dir      string
        interval time.Duration
        hidden   bool
        p        string
)

func loop() {
        watcher, err := fsnotify.NewWatcher()
        if err != nil {
                log.Fatal(err)
        }
        err = watcher.Add(dir)
        if err != nil {
                log.Fatal(err)
        }
        ready := make(chan bool, 1)

        // run commands at interval
        go func() {
                wait := time.Tick(interval)
                for _ = range wait {
                        select {
                        case <-ready:
                                go run()
                        default:
                        }
                }
        }()

        // filter events
        for {
                select {
                case ev := <-watcher.Events:
                        if ev.Op&fsnotify.Write != fsnotify.Write {
                                break
                        }
                        if err := watched(ev.Name); err != nil {
                                break
                        }
                        // name, _ := filepath.Rel(dir, ev.Name)
                        select {
                        case ready <- true:
                        default:
                        }
                case err := <-watcher.Errors:
                        log.Println("ift error:", err)
                }
        }
}

func run() {
        c := exec.Command("sh", "-c", cmd)
        c.Dir = dir
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr
        if err := c.Run(); err != nil {
                log.Println(err)
        }
}

func watched(path string) error {
        if path == "" {
                return errors.New("file name not found")
        }
        if hidden && filepath.Base(path)[0] == '.' {
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
        flag.DurationVar(&interval, "interval", 2*time.Second, "Interval seconds")
        flag.BoolVar(&hidden, "hidden", false, "Watch hidden file")
        flag.StringVar(&p, "p", "", "Specify file name patterns to watch. "+
                "Multiple patterns should be seperated by comma. "+
                "If pattern is not specified all files in the dir will be watched(except hidden files)")
        flag.Parse()

        dir, _ = filepath.Abs(dir)
        pats := strings.Split(p, ",")
        for _, pat := range pats {
                patterns = append(patterns, strings.TrimSpace(pat))
        }

        if flag.NArg() != 1 {
                fmt.Println("Command not found")
                os.Exit(1)
        }

        cmd = flag.Arg(0)
        loop()
}
