package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

type Client struct {
	bot     *tgbotapi.BotAPI
	groupID int64
	cache   *TopicCache
}

func NewClient(token string, groupID int64, dataDir string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("初始化Telegram Bot失败: %w", err)
	}

	cache, err := NewTopicCache(dataDir)
	if err != nil {
		return nil, fmt.Errorf("初始化 Topic 缓存失败: %w", err)
	}

	return &Client{
		bot:     bot,
		groupID: groupID,
		cache:   cache,
	}, nil
}

// getOrCreateTopic 获取或创建topic
func (c *Client) getOrCreateTopic(sender string) (int, error) {
	// 先从缓存中查找
	if topicID, exists := c.cache.GetTopicID(c.groupID, sender); exists {
		return topicID, nil
	}

	// 创建新topic
	createConfig := tgbotapi.CreateForumTopicConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: c.groupID},
		Name:       sender,
	}

	msg, err := c.bot.Send(createConfig)
	if err != nil {
		return 0, fmt.Errorf("创建topic失败: %w", err)
	}

	// 保存到缓存
	topicID := msg.MessageThreadID
	if err := c.cache.SetTopicID(c.groupID, sender, topicID); err != nil {
		log.Printf("警告: 保存topic缓存失败: %v", err)
	}

	return topicID, nil
}

// SendMessage 发送消息到指定sender对应的topic
func (c *Client) SendMessage(sender, text string) error {
	topicID, err := c.getOrCreateTopic(sender)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(c.groupID, text)
	msg.MessageThreadID = topicID

	_, err = c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}
