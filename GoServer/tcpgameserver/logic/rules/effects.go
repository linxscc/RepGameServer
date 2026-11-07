package rules

// This file is the registry of every behaviour the yaml tables may reference.
// Numbers live in the tables; what a key *does* lives in code (score.go,
// economy.go, merge.go). Validation rejects unknown keys and missing params.

// bondModifiers are the generic per-tier values a bond may set.
// scoring = evaluated on the board being scored; ops = on the last settled board.
var bondModifiers = map[string]struct{ ops bool }{
	BondPointsPerCard: {false}, // 每张出阵卡 +点
	BondAddMult:       {false}, // +倍率
	BondTimesMult:     {false}, // ×倍率
}

// legendParams: 3★ skill key → required params.
var legendParams = map[string][]string{
	LegendMultPerHighStar: {"mult"},        // 百工专家：每张 2★+ 出阵卡 +倍率
	LegendGrowOnMerge:     {"points"},      // 铸剑师：每次合成永久 +点
	LegendOwnOriginCount:  {"count"},       // 神农：自己的出身羁绊计数 +N
	LegendRetriggerLeft:   {"count"},       // 天师：最左边法宝额外触发 N 次
	LegendMultPerRelic:    {"mult"},        // 住持：每个法宝 +倍率
	LegendMultPerLevel:    {"mult"},        // 教授：每级 +倍率
	LegendDamageTaken:     {"factor"},      // 医圣：受到伤害 ×factor
	LegendMultPerBond:     {"mult"},        // 艺术家：每个生效羁绊 +倍率
	LegendMultPerLine:     {"mult"},        // 营造宗师：每个不同的系 +倍率
	LegendInterest:        {"rate", "cap"}, // 商人：利息每 step +rate，上限 cap
	LegendTimesMult:       {"factor"},      // 统领：倍率 ×factor
	LegendExtraSlots:      {"slots"},       // 相国：出阵格 +N
}

// legend skills that only work on the last settled board (like ops bonds).
var opsLegend = map[string]bool{LegendInterest: true, LegendExtraSlots: true}

type relicKind int

const (
	kindPerCard relicKind = iota // 逐卡
	kindBench                    // 备战
	kindWhole                    // 整体
	kindOps                      // 运营
	kindCopy                     // 复制
	kindCount                    // 计数
	kindFactor                   // 逐卡倍率：本卡当前战力 ×factor（逐卡管线中先于其余逐卡法宝）
	kindFoeCard                  // 特攻：在对手的逐卡结算里生效（每张卡每次结算 1 次，不受重复触发影响）
)

type relicBehaviour struct {
	kind   relicKind
	params []string
	growth bool // 记录成长值
	times  bool // ×倍率：额外触发时只生效一次
	choice bool // 获得时选一个行当
}

var relicBehaviours = map[string]relicBehaviour{
	RelicFurnace:      {kind: kindWhole, params: []string{"per_merge"}, growth: true},
	RelicHeirloom:     {kind: kindPerCard, params: []string{"mult"}},
	RelicMasterSeal:   {kind: kindWhole, params: []string{"factor"}, times: true},
	RelicGraduation:   {kind: kindOps, params: []string{"gold"}},
	RelicFellowship:   {kind: kindWhole, params: []string{"mult"}},
	RelicSectToken:    {kind: kindCount, params: []string{"count", "mult"}, choice: true},
	RelicEclectic:     {kind: kindWhole, params: []string{"min_lines", "factor"}, times: true},
	RelicSecretArt:    {kind: kindPerCard, params: []string{"repeats"}, choice: true},
	RelicLoneWolf:     {kind: kindWhole, params: []string{"max_cards", "factor"}, times: true},
	RelicLean:         {kind: kindWhole, params: []string{"mult"}},
	RelicVanguard:     {kind: kindPerCard, params: []string{"factor"}},
	RelicMultitude:    {kind: kindWhole, params: []string{"mult"}},
	RelicNewWave:      {kind: kindPerCard, params: []string{"points", "mult"}},
	RelicRearguard:    {kind: kindPerCard, params: []string{"mult"}},
	RelicHidden:       {kind: kindBench, params: []string{"mult"}},
	RelicStorehouse:   {kind: kindBench, params: []string{"points"}},
	RelicTreasure:     {kind: kindOps, params: []string{"cap"}},
	RelicPiggy:        {kind: kindWhole, params: []string{"per_gold", "mult"}},
	RelicThrifty:      {kind: kindOps, params: []string{"gold"}},
	RelicAccountant:   {kind: kindWhole, params: []string{"per_sell"}, growth: true},
	RelicGrudge:       {kind: kindWhole, params: []string{"per_loss", "cap"}, growth: true},
	RelicMomentum:     {kind: kindWhole, params: []string{"mult"}},
	RelicLastStand:    {kind: kindWhole, params: []string{"hp_below", "factor"}, times: true},
	RelicTailwind:     {kind: kindWhole, params: []string{"per_level"}, times: true},
	RelicLiteraryPen:  {kind: kindWhole, params: []string{"factor"}, times: true},
	RelicMarketPass:   {kind: kindOps, params: []string{"free"}},
	RelicAcademy:      {kind: kindOps, params: []string{"xp"}},
	RelicExpansion:    {kind: kindOps, params: []string{"slots"}},
	RelicTempering:    {kind: kindOps, params: []string{"points"}},
	RelicCinnabar:     {kind: kindPerCard, params: []string{"mult"}},
	RelicPineInk:      {kind: kindPerCard, params: []string{"points"}},
	RelicShrine:       {kind: kindFactor, params: []string{"factor"}},
	RelicNightParade:  {kind: kindWhole, params: []string{"min_cards", "factor"}, times: true},
	RelicOneColor:     {kind: kindWhole, params: []string{"min_cards", "factor"}, times: true},
	RelicYinYang:      {kind: kindWhole, params: []string{"min_each", "factor"}, times: true},
	RelicFourSymbols:  {kind: kindWhole, params: []string{"mult"}},
	RelicVermilion:    {kind: kindOps},
	RelicDemonMirror:  {kind: kindFoeCard, params: []string{"factor"}},
	RelicGodBreaker:   {kind: kindFoeCard, params: []string{"factor"}},
	RelicBehead:       {kind: kindFoeCard},
	RelicColorCounter: {kind: kindWhole, params: []string{"mult"}},
	RelicUsurp:        {kind: kindWhole, params: []string{"mult"}},
	RelicDefyHeaven:   {kind: kindWhole, params: []string{"factor"}, times: true},
	RelicMirror:       {kind: kindCopy},
	RelicRubbing:      {kind: kindCopy},
}

var relicParams = func() map[string][]string {
	m := map[string][]string{}
	for k, b := range relicBehaviours {
		m[k] = b.params
	}
	return m
}()

// omenParams: omen key → required params.
var omenParams = map[string][]string{
	OmenQuake:        {"bench"},
	OmenSolarEclipse: {"factor"},
	OmenGhostRout:    {"factor"},
	OmenRedMoon:      {"factor"},
	OmenInkRain:      {"factor"},
	OmenFamine:       {"income"},
	OmenMarketClosed: {"refresh"},
	OmenHeavyTax:     {},
	OmenShortage:     {"shop"},
	OmenNarrowPath:   {"slots"},
	OmenDiscord:      {"at"},
	OmenUpheaval:     {"factor"},
}

// familyParams: family key → required params.
var familyParams = map[string][]string{
	FamilyMerchant: {"gold"},                         // 开局 +金
	FamilyScholar:  {"relics"},                       // 开局获得 N 个最低稀有度法宝
	FamilyGeneral:  {"level"},                        // 开局等级
	FamilyArtisan:  {"star"},                         // 开局第一张卡的星级
	FamilyNoble:    {"extra_offer"},                  // 宝库多 N 个选项
	FamilyPoor:     {"gold_per_loss"},                // 每输一回合额外 +金
	FamilyMiser:    {"interest_cap", "income_delta"}, // 利息上限 +N，基础收入 +delta
	FamilyHermit:   {"relic_slots", "board_slots"},   // 法宝栏 +N，出阵格 +N（可为负）
}
