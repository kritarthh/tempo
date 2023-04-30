package main

import (
    "fmt"
    "github.com/kataras/golog"
    "runtime"
    "strings"
)

func GetLogger(id string, callerSkip int) *golog.Logger {
    newLogger := golog.New()
    newLogger.SetLevel("debug")
    newLogger.Handle(func(l *golog.Log) bool {
        prefix := golog.GetTextForLevel(l.Level, true)
        _, fn, line, _ := runtime.Caller(callerSkip)
        message := fmt.Sprintf("%s %s %s:%d %s: %s",
            l.Time.UTC(), id, fn[strings.LastIndex(fn, "/")+1:], line, prefix, l.Message)

        if l.NewLine {
            message += "\n"
        }

        fmt.Print(message)
        return true
    })
    return newLogger
}
