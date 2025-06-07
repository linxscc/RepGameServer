package models

import (
	"fmt"
	"sync/atomic"
)

// Card 卡牌模型，存储在共享池中的值类型
type Card struct {
	UID        string  `json:"UID"`        // 卡牌唯一实例ID，实例化时赋值
	ID         int     `json:"ID"`         // 卡牌ID，数据库中唯一标识
	Name       string  `json:"Name"`       // 卡牌名称
	Damage     float64 `json:"Damage"`     // 伤害值
	TargetName *string `json:"TargetName"` // 目标名称（可能为空）
	Level      int     `json:"Level"`      // 卡牌等级
}

// 全局UID计数器，用于生成唯一的卡牌实例ID
var cardUIDCounter int64

// NewCard 创建新的卡牌实例
func NewCard(id int, name string, damage float64, targetName *string, level int) Card {
	counter := atomic.AddInt64(&cardUIDCounter, 1)
	uid := fmt.Sprintf("card_%d_%d", id, counter)
	return Card{
		UID:        uid,
		ID:         id,
		Name:       name,
		Damage:     damage,
		TargetName: targetName,
		Level:      level,
	}
}

// String 返回卡牌的字符串表示
func (c Card) String() string {
	targetStr := "无目标"
	if c.TargetName != nil {
		targetStr = *c.TargetName
	}
	return fmt.Sprintf("Card[UID:%s, Name:%s, Damage:%.2f, Target:%s, Level:%d]",
		c.UID, c.Name, c.Damage, targetStr, c.Level)
}
