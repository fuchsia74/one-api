package message

import (
	"fmt"

	"github.com/Laisky/errors/v2"

	"github.com/songquanpeng/one-api/common/config"
)

const (
	ByAll           = "all"
	ByEmail         = "email"
	ByMessagePusher = "message_pusher"
)

func Notify(by string, title string, description string, content string) error {
	switch by {
	case ByAll:
		var errMsgs []string
		if err := SendEmail(title, config.RootUserEmail, content); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("failed to send email: %v", err))
		}
		if err := SendMessage(title, description, content); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("failed to send message: %v", err))
		}

		if len(errMsgs) > 0 {
			return errors.Errorf("multiple errors occurred: %v", errMsgs)
		}
		return nil
	case ByEmail:
		return SendEmail(title, config.RootUserEmail, content)
	case ByMessagePusher:
		return SendMessage(title, description, content)
	default:
		return errors.Errorf("unknown notify method: %s", by)
	}
}
