package main

import (
        "errors"
        "flag"
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "time"
)

var (
        d         string
        cmd       string
        updatedAt time.Time
        n         float64
)

func loop() {
        updatedAt = time.Now()
        for {
                filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
                        if err != nil {
                                fmt.Fprint(os.Stderr, err)
                                return err
                        }

                        if info.ModTime().After(updatedAt) {
                                fmt.Println(path, "changed. Running command", cmd)
                                updatedAt = time.Now()
                                c := exec.Command("sh", "-c", cmd)
                                c.Stdout = os.Stdout
                                c.Stderr = os.Stderr
                                if err := c.Run(); err != nil {
                                        fmt.Fprint(os.Stderr, err)
                                }
                                return errors.New("done")
                        }

                        return nil
                })
                time.Sleep(time.Duration(n) * time.Second)
        }
}

func main() {
        flag.Usage = func() {
                fmt.Println("ift [-d dir] [-n secs] command")
                fmt.Println("\nOPTIONS:")
                flag.PrintDefaults()
        }
        flag.StringVar(&d, "d", ".", "watch directory")
        flag.Float64Var(&n, "n", 1.0, "interval seconds")
        flag.Parse()
        d, _ = filepath.Abs(d)
        if flag.NArg() != 1 {
                fmt.Println("Command not found")
                os.Exit(1)
        }
        cmd = flag.Arg(0)
        loop()
}
