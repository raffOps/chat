package chat

import (
	"github.com/raffops/chat_commons/pkg/errs"
)

type Repository interface {
	GetChannel(queueId string) (chan map[string]interface{}, errs.ChatError)
}
