package events

import (
	"fmt"
	"log"
	"sync"
)

// EventHandler 事件处理器类型
type EventHandler func(data interface{})

// EventSubscription 事件订阅信息
type EventSubscription struct {
	ID       string       // 订阅ID
	Handler  EventHandler // 处理函数
	Priority int          // 优先级，数字越小优先级越高
}

// EventManager 事件管理器
type EventManager struct {
	subscribers map[string][]*EventSubscription // 事件类型 -> 订阅者列表
	mutex       sync.RWMutex                    // 读写锁
	idCounter   int                             // 订阅ID计数器
}

var (
	globalEventManager *EventManager
	once               sync.Once
)

// GetEventManager 获取全局事件管理器单例
func GetEventManager() *EventManager {
	once.Do(func() {
		globalEventManager = &EventManager{
			subscribers: make(map[string][]*EventSubscription),
			idCounter:   0,
		}
	})
	return globalEventManager
}

// Subscribe 订阅事件
// eventType: 事件类型
// handler: 事件处理函数
// priority: 优先级（可选，默认为100）
func (em *EventManager) Subscribe(eventType string, handler EventHandler, priority ...int) string {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// 生成唯一订阅ID
	em.idCounter++
	subscriptionID := fmt.Sprintf("%s_%d", eventType, em.idCounter)

	// 设置优先级
	prio := 100
	if len(priority) > 0 {
		prio = priority[0]
	}

	// 创建订阅
	subscription := &EventSubscription{
		ID:       subscriptionID,
		Handler:  handler,
		Priority: prio,
	}

	// 添加到订阅列表
	if _, exists := em.subscribers[eventType]; !exists {
		em.subscribers[eventType] = make([]*EventSubscription, 0)
	}

	em.subscribers[eventType] = append(em.subscribers[eventType], subscription)

	// 按优先级排序（数字越小优先级越高）
	em.sortSubscriptionsByPriority(eventType)

	log.Printf("Event subscribed: %s, ID: %s, Priority: %d", eventType, subscriptionID, prio)
	return subscriptionID
}

// Unsubscribe 取消订阅事件
func (em *EventManager) Unsubscribe(subscriptionID string) bool {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	for eventType, subscriptions := range em.subscribers {
		for i, subscription := range subscriptions {
			if subscription.ID == subscriptionID {
				// 移除订阅
				em.subscribers[eventType] = append(subscriptions[:i], subscriptions[i+1:]...)
				log.Printf("Event unsubscribed: %s, ID: %s", eventType, subscriptionID)
				return true
			}
		}
	}

	log.Printf("Subscription not found: %s", subscriptionID)
	return false
}

// UnsubscribeAll 取消指定事件类型的所有订阅
func (em *EventManager) UnsubscribeAll(eventType string) int {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	count := 0
	if subscriptions, exists := em.subscribers[eventType]; exists {
		count = len(subscriptions)
		delete(em.subscribers, eventType)
		log.Printf("All subscriptions removed for event type: %s, count: %d", eventType, count)
	}

	return count
}

// Publish 发布事件
func (em *EventManager) Publish(eventType string, data interface{}) {
	em.mutex.RLock()
	subscriptions := make([]*EventSubscription, 0)
	if subs, exists := em.subscribers[eventType]; exists {
		// 创建副本避免并发问题
		subscriptions = make([]*EventSubscription, len(subs))
		copy(subscriptions, subs)
	}
	em.mutex.RUnlock()

	if len(subscriptions) == 0 {
		log.Printf("No subscribers for event: %s", eventType)
		return
	}

	// 执行所有订阅者的处理函数
	for _, subscription := range subscriptions {
		go func(sub *EventSubscription) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Event handler panic for %s (ID: %s): %v", eventType, sub.ID, r)
				}
			}()

			sub.Handler(data)
		}(subscription)
	}
}

// PublishSync 同步发布事件（等待所有处理器完成）
func (em *EventManager) PublishSync(eventType string, data interface{}) {
	em.mutex.RLock()
	subscriptions := make([]*EventSubscription, 0)
	if subs, exists := em.subscribers[eventType]; exists {
		subscriptions = make([]*EventSubscription, len(subs))
		copy(subscriptions, subs)
	}
	em.mutex.RUnlock()

	if len(subscriptions) == 0 {
		log.Printf("No subscribers for event: %s", eventType)
		return
	}

	log.Printf("Publishing sync event: %s to %d subscribers", eventType, len(subscriptions))

	var wg sync.WaitGroup
	for _, subscription := range subscriptions {
		wg.Add(1)
		go func(sub *EventSubscription) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Event handler panic for %s (ID: %s): %v", eventType, sub.ID, r)
				}
			}()

			sub.Handler(data)
		}(subscription)
	}

	wg.Wait()
	log.Printf("Sync event completed: %s", eventType)
}

// GetSubscriberCount 获取指定事件类型的订阅者数量
func (em *EventManager) GetSubscriberCount(eventType string) int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if subscriptions, exists := em.subscribers[eventType]; exists {
		return len(subscriptions)
	}
	return 0
}

// GetAllEventTypes 获取所有事件类型
func (em *EventManager) GetAllEventTypes() []string {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	eventTypes := make([]string, 0, len(em.subscribers))
	for eventType := range em.subscribers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// GetEventStats 获取事件统计信息
func (em *EventManager) GetEventStats() map[string]interface{} {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	stats := make(map[string]interface{})
	totalSubscribers := 0

	for eventType, subscriptions := range em.subscribers {
		stats[eventType] = len(subscriptions)
		totalSubscribers += len(subscriptions)
	}

	stats["total_event_types"] = len(em.subscribers)
	stats["total_subscribers"] = totalSubscribers

	return stats
}

// sortSubscriptionsByPriority 按优先级排序订阅
func (em *EventManager) sortSubscriptionsByPriority(eventType string) {
	if subscriptions, exists := em.subscribers[eventType]; exists {
		// 使用简单的冒泡排序按优先级排序
		n := len(subscriptions)
		for i := 0; i < n-1; i++ {
			for j := 0; j < n-i-1; j++ {
				if subscriptions[j].Priority > subscriptions[j+1].Priority {
					subscriptions[j], subscriptions[j+1] = subscriptions[j+1], subscriptions[j]
				}
			}
		}
	}
}

// HasSubscribers 检查是否有订阅者
func (em *EventManager) HasSubscribers(eventType string) bool {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if subscriptions, exists := em.subscribers[eventType]; exists {
		return len(subscriptions) > 0
	}
	return false
}

// Clear 清空所有订阅
func (em *EventManager) Clear() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	totalSubscribers := 0
	for _, subscriptions := range em.subscribers {
		totalSubscribers += len(subscriptions)
	}

	em.subscribers = make(map[string][]*EventSubscription)
	em.idCounter = 0

	log.Printf("All event subscriptions cleared, total removed: %d", totalSubscribers)
}
