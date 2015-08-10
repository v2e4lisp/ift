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
        ignorefile string
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

        filepath.Walk(dir, func (path string, info os.FileInfo, err error) error {
                if !info.IsDir() {
                        return nil
                }

                if err != nil {
                        log.Println(err)
                        return nil
                }

                if err := watched(path); err != nil {
                        return filepath.SkipDir
                }

                log.Println("watch: ", path)
                watcher.Add(path)
                return nil
        });

        nextRound := time.Now()
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
                        now := time.Now()
                        if !now.After(nextRound) {
                                break
                        }
                        nextRound = now.Add(interval)

                        if wait {
                                run(&ev)
                        } else {
                                go run(&ev)
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
                return errors.New("file ignored");
        }
        return nil
}

func loadIgnoreFile() error {
        text, err := ioutil.ReadFile(ignorefile)
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
                fmt.Println("  ift [-d dir] [-ignorefile path] [-n interval] [-p patterns] [-wait] [-hidden] command")
                fmt.Println("\nOptions:")
                flag.PrintDefaults()
        }
        flag.StringVar(&dir, "d", ".", "Watch directory")
        flag.DurationVar(&interval, "n", 2*time.Second, "Interval between command execution")
        flag.BoolVar(&hidden, "hidden", false, "Watch hidden file")
        flag.BoolVar(&wait, "wait", false, "Wait for last command to finish.")

        flag.StringVar(&ignorefile, "ignorefile", ".iftignore", "contains file patterns to ignore. "+
                "ift use these patterns to determine which files to ignore. "+
                "If ignorefile is not specified, ift will try to load "+
                ".iftignore file under the watch directory. "+
                "You can also specify patterns using -p option. ")

        flag.StringVar(&p, "p", "", "Specify file name patterns to ignore. "+
                "Multiple patterns should be seperated by comma. "+
                "If pattern is not specified, "+
                "all files in the dir will be watched(except hidden files). "+
                "You can also use ignore file to specify patterns.")
        flag.Parse()
        if flag.NArg() < 1 {
                flag.Usage()
                os.Exit(1)
        }

        dir, _ = filepath.Abs(dir)
        if !filepath.IsAbs(ignorefile) {
                ignorefile = filepath.Join(dir, ignorefile)
        }

        loadIgnoreFile()
        if p != "" {
                pats := strings.Split(p, ",")
                for _, pat := range pats {
                        patterns = append(patterns, strings.TrimSpace(pat))
                }
        }

        cmd = strings.Join(flag.Args(), " ")
        loop()
}
