package events

// 全局便捷函数，直接使用全局事件管理器

// Subscribe 订阅事件（全局函数）
func Subscribe(eventType string, handler EventHandler, priority ...int) string {
	return GetEventManager().Subscribe(eventType, handler, priority...)
}

// Unsubscribe 取消订阅事件（全局函数）
func Unsubscribe(subscriptionID string) bool {
	return GetEventManager().Unsubscribe(subscriptionID)
}

// UnsubscribeAll 取消指定事件类型的所有订阅（全局函数）
func UnsubscribeAll(eventType string) int {
	return GetEventManager().UnsubscribeAll(eventType)
}

// Publish 发布事件（全局函数）
func Publish(eventType string, data interface{}) {
	GetEventManager().Publish(eventType, data)
}

// PublishSync 同步发布事件（全局函数）
func PublishSync(eventType string, data interface{}) {
	GetEventManager().PublishSync(eventType, data)
}

// GetSubscriberCount 获取指定事件类型的订阅者数量（全局函数）
func GetSubscriberCount(eventType string) int {
	return GetEventManager().GetSubscriberCount(eventType)
}

// GetAllEventTypes 获取所有事件类型（全局函数）
func GetAllEventTypes() []string {
	return GetEventManager().GetAllEventTypes()
}

// GetEventStats 获取事件统计信息（全局函数）
func GetEventStats() map[string]interface{} {
	return GetEventManager().GetEventStats()
}

// HasSubscribers 检查是否有订阅者（全局函数）
func HasSubscribers(eventType string) bool {
	return GetEventManager().HasSubscribers(eventType)
}

// Clear 清空所有订阅（全局函数）
func Clear() {
	GetEventManager().Clear()
}
