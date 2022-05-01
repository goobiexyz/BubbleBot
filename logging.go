package bubble

import (
  "fmt"
  "log"
  "github.com/fatih/color"
)

type LogStatus int
const (
  Info LogStatus = iota
  Warning
  Error
  lifecycle
)

func Log(status LogStatus, id, details string) {
  msg := LogStr(status, id, details)
  log.Print(msg)
}


func LogStr(status LogStatus, id, details string) string {
  msg := fmt.Sprintf("[%s] %s", id, details)

  switch status {
  case Info:
    // use default color
  case Warning:
    msg = color.YellowString(msg)
  case Error:
    msg = color.RedString(msg)
  case lifecycle:
    msg = color.GreenString(msg)
  }

  return msg
}
