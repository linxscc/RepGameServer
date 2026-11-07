package chessbot

// 机器人决策用的经验阈值：只影响机器人怎么玩，不影响对局规则。
const (
	maxRefreshesPerTurn = 12 // 每个准备阶段最多刷新商店次数（防止死循环）
	earlyRefreshRounds  = 4  // 第 4 回合之前不花钱刷新商店，先攒经济
	refreshGoldBuffer   = 2  // 花钱刷新后，金币至少还要比储备金多这么多
	lowHP               = 35 // 血量 ≤ 该值时不再存钱（储备金视为 0），全力提升战力
	keepCardWorth       = 20 // 卡牌估值 ≥ 该值时，不为腾位置卖掉它
	openingRounds       = 3  // 前 3 回合见到最低费用的卡就买
	affordableCost      = 2  // 凑齐合成所需的卡时，费用 ≤ 该值才买
	arrangePasses       = 2  // 布阵时"出阵卡选择"与"法宝顺序"交替优化的轮数
)
