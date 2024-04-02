package handle_defender

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/umfaka/tgfaka/internal/exts/tg_bot"
	"github.com/umfaka/tgfaka/internal/log"
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
	log.Error(msgText)
	tg_bot.SendAdmin(msgText)
}

func HandleError(err error, prefixString string) {
	msgText := fmt.Sprintf("%s, Error: %v", prefixString, err)
	log.Error(msgText)
	tg_bot.SendAdmin(msgText)
}
