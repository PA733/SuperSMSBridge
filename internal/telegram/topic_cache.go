package telegram

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type TopicCache struct {
	GroupTopics map[int64]map[string]int `json:"group_topics"` // groupID -> sender -> topicID
	mutex       sync.RWMutex
	filePath    string
}

func NewTopicCache(cacheDir string) (*TopicCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	filePath := filepath.Join(cacheDir, "topic_cache.json")
	cache := &TopicCache{
		GroupTopics: make(map[int64]map[string]int),
		filePath:    filePath,
	}

	// 尝试加载现有缓存
	if err := cache.load(); err != nil {
		// 如果文件不存在，使用空缓存
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载缓存失败: %w", err)
		}
	}

	return cache, nil
}

func (c *TopicCache) GetTopicID(groupID int64, sender string) (int, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if topics, ok := c.GroupTopics[groupID]; ok {
		if topicID, exists := topics[sender]; exists {
			return topicID, true
		}
	}
	return 0, false
}

func (c *TopicCache) SetTopicID(groupID int64, sender string, topicID int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.GroupTopics[groupID]; !ok {
		c.GroupTopics[groupID] = make(map[string]int)
	}
	c.GroupTopics[groupID][sender] = topicID

	return c.save()
}

func (c *TopicCache) load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	return json.Unmarshal(data, &c.GroupTopics)
}

func (c *TopicCache) save() error {
	data, err := json.MarshalIndent(c.GroupTopics, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	return os.WriteFile(c.filePath, data, 0644)
}
