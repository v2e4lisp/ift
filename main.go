package main

import (
        "errors"
        "flag"
        "fmt"
        "io/ioutil"
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
        dir       string
        interval  time.Duration
        watchfile string
        wait      bool
        hidden    bool
        p         string
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
        ready := make(chan *fsnotify.Event, 1)
        go func() {
                t := make(chan interface{})
                for {
                        ev := <-ready
                        go func() { time.Sleep(interval); t <- true }()
                        if wait {
                                run(ev)
                        } else {
                                go run(ev)
                        }
                        <-t
                }
        }()

        // filter events
        for {
                select {
                case ev := <-watcher.Events:
                        // ignore chmod
                        if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
                                break
                        }
                        if err := watched(ev.Name); err != nil {
                                break
                        }
                        select {
                        case ready <- &ev:
                        default:
                        }
                case err := <-watcher.Errors:
                        log.Println("ift error:", err)
                }
        }
}

func run(ev *fsnotify.Event) {
        log.Println("sh -c", cmd)
        env := append(os.Environ(), fmt.Sprintf("FS_EVENT=%s", ev))
        c := exec.Command("sh", "-c", cmd)
        c.Env = env
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
        if !hidden && filepath.Base(path)[0] == '.' {
                return errors.New("hidden file")
        }
        // watch all
        if len(patterns) == 0 {
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

func loadWatchFile() error {
        text, err := ioutil.ReadFile(watchfile)
        if err != nil {
                return err
        }
        lines := strings.Split(string(text), "\n")
        for _, l := range lines {
                l = strings.TrimSpace(l)
                if l == "" || l[0] == '#' {
                        continue
                }
                patterns = append(patterns, l)
        }
        return nil
}

func main() {
        flag.Usage = func() {
                fmt.Println("Usage:")
                fmt.Println("  ift [-d dir] [-watchfile path] [-n interval] [-p patterns] [-wait] [-hidden] command")
                fmt.Println("\nOptions:")
                flag.PrintDefaults()
        }
        flag.StringVar(&dir, "d", ".", "Watch directory")
        flag.DurationVar(&interval, "n", 2*time.Second, "Interval between command execution")
        flag.BoolVar(&hidden, "hidden", false, "Watch hidden file")
        flag.BoolVar(&wait, "wait", false, "Wait for last command to finish.")

        flag.StringVar(&watchfile, "watchfile", ".watch", "Watch file contains file name patterns. "+
                "ift use these patterns to determine which files to watch. "+
                "If watchfile is not specified, ift will try to load "+
                ".watch file under the watch directory. "+
                "You can also specify patterns using -p option. ")

        flag.StringVar(&p, "p", "", "Specify file name patterns to watch. "+
                "Multiple patterns should be seperated by comma. "+
                "If pattern is not specified, "+
                "all files in the dir will be watched(except hidden files). "+
                "You can also use watch file to specify patterns.")
        flag.Parse()
        if flag.NArg() < 1 {
                flag.Usage()
                os.Exit(1)
        }

        dir, _ = filepath.Abs(dir)
        if !filepath.IsAbs(watchfile) {
                watchfile = filepath.Join(dir, watchfile)
        }
        loadWatchFile()
        if p != "" {
                pats := strings.Split(p, ",")
                for _, pat := range pats {
                        patterns = append(patterns, strings.TrimSpace(pat))
                }
        }

        cmd = strings.Join(flag.Args(), " ")
        loop()
}
