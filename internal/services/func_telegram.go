package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/umfaka/tgfaka/internal/exts/tg_bot"
	"github.com/umfaka/tgfaka/internal/log"
)

// 要求机器人具有管理员权限
func createInvite(chatIDStr string, memberLimit int) (string, error) {
	// 设置创建群组邀请链接的参数配置
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("strconv.ParseInt(%s): %v", chatIDStr, err)
	}
	inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
		// ExpireDate:         3600,  // 过期时间（秒）
		MemberLimit:        memberLimit, // 最大成员限制
		CreatesJoinRequest: false,       // 是否需要管理员确认
	}

	// 创建群组邀请链接
	resp, err := tg_bot.Bot.Request(inviteLinkConfig)
	if err != nil {
		// log.Fatal(err)
		log.Errorf("Bot.Request(%+v): %v", inviteLinkConfig, err)
		return "", fmt.Errorf("Bot.Request: %v", err)
	}

	inviteLink := &tgbotapi.ChatInviteLink{}
	invite := resp.Result

	err = json.Unmarshal(invite, inviteLink)
	if err != nil {
		log.Errorf("Bot.Request(%+v): %v", invite, err)
		return "", fmt.Errorf("json.Unmarshal: %v", err)
	}
	return inviteLink.InviteLink, nil
}
