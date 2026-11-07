package rules

// 卡牌星级与费用档。合成：merge.group_size 张同系同星合成一张高一星的卡，
// 最高星级以配置 merge.max_star 为准（当前为 3）。
const (
	StarOne   = 1 // 1★：商店买到的初始卡
	StarTwo   = 2 // 2★：同系 1★ 合成而来
	StarThree = 3 // 3★：同系 2★ 合成而来，解锁该系的 3★ 传奇技能

	CostCheapest = 1 // 最低费用档：开局卡所在档（chess_rules.yaml start_card_cost）
)
