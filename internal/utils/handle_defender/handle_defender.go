package handle_defender

import (
	"fmt"
	my_log "gopay/internal/exts/log"
	"gopay/internal/exts/tg_bot"
	"runtime"
	"strings"
)

func HandlePanic(r interface{}, prefixString string) {
	var msg string
	threshold := 2
	for skip, stackNum := 0, 1; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			msg = fmt.Sprintf("Unable to retrieve panic information.")
			break
		}

		funcName := runtime.FuncForPC(pc).Name()
		if !strings.Contains(funcName, "runtime.") && !strings.Contains(file, "handle_defender.go") {
			msg = msg + fmt.Sprintf("\nFunction: %s, File: %s, Line: %d, Panic: %v", funcName, file, line, r)

			stackNum = stackNum + 1
			if stackNum > threshold {
				break
			}
		}
	}

	msgText := fmt.Sprintf("%s, Error: %s", prefixString, msg)
	my_log.LogError(msgText)
	tg_bot.SendAdmin(msgText)
}

func HandleError(err error, prefixString string) {
	msgText := fmt.Sprintf("%s, Error: %v", prefixString, err)
	my_log.LogError(msgText)
	tg_bot.SendAdmin(msgText)
}
