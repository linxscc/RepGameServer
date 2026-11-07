package rules

// This file is the registry of every behaviour the yaml tables may reference.
// Numbers live in the tables; what a key *does* lives in code (score.go,
// economy.go, merge.go). Validation rejects unknown keys and missing params.

// bondModifiers are the generic per-tier values a bond may set.
// scoring = evaluated on the board being scored; ops = on the last settled board.
var bondModifiers = map[string]struct{ ops bool }{
	"points_per_card":     {false}, // 每张出阵卡 +点
	"add_mult":            {false}, // +倍率
	"times_mult":          {false}, // ×倍率
}

// legendParams: 3★ skill key → required params.
var legendParams = map[string][]string{
	"mult_per_high_star": {"mult"},        // 百工专家：每张 2★+ 出阵卡 +倍率
	"grow_on_merge":      {"points"},      // 铸剑师：每次合成永久 +点
	"own_origin_count":   {"count"},       // 神农：自己的出身羁绊计数 +N
	"retrigger_left":     {"count"},       // 天师：最左边法宝额外触发 N 次
	"mult_per_relic":     {"mult"},        // 住持：每个法宝 +倍率
	"mult_per_level":     {"mult"},        // 教授：每级 +倍率
	"damage_taken":       {"factor"},      // 医圣：受到伤害 ×factor
	"mult_per_bond":      {"mult"},        // 艺术家：每个生效羁绊 +倍率
	"mult_per_line":      {"mult"},        // 营造宗师：每个不同的系 +倍率
	"interest":           {"rate", "cap"}, // 商人：利息每 step +rate，上限 cap
	"times_mult":         {"factor"},      // 统领：倍率 ×factor
	"extra_slots":        {"slots"},       // 相国：出阵格 +N
}

// legend skills that only work on the last settled board (like ops bonds).
var opsLegend = map[string]bool{"interest": true, "extra_slots": true}

type relicKind int

const (
	kindPerCard relicKind = iota // 逐卡
	kindBench                    // 备战
	kindWhole                    // 整体
	kindOps                      // 运营
	kindCopy                     // 复制
	kindCount                    // 计数
	kindFactor                   // 逐卡倍率：本卡当前战力 ×factor（逐卡管线中先于其余逐卡法宝）
)

type relicBehaviour struct {
	kind   relicKind
	params []string
	growth bool // 记录成长值
	times  bool // ×倍率：额外触发时只生效一次
	choice bool // 获得时选一个行当
}

var relicBehaviours = map[string]relicBehaviour{
	"furnace":      {kind: kindWhole, params: []string{"per_merge"}, growth: true},
	"heirloom":     {kind: kindPerCard, params: []string{"mult"}},
	"master_seal":  {kind: kindWhole, params: []string{"factor"}, times: true},
	"graduation":   {kind: kindOps, params: []string{"gold"}},
	"fellowship":   {kind: kindWhole, params: []string{"mult"}},
	"sect_token":   {kind: kindCount, params: []string{"count", "mult"}, choice: true},
	"eclectic":     {kind: kindWhole, params: []string{"min_lines", "factor"}, times: true},
	"secret_art":   {kind: kindPerCard, params: []string{"repeats"}, choice: true},
	"lone_wolf":    {kind: kindWhole, params: []string{"max_cards", "factor"}, times: true},
	"lean":         {kind: kindWhole, params: []string{"mult"}},
	"vanguard":     {kind: kindPerCard, params: []string{"factor"}},
	"multitude":    {kind: kindWhole, params: []string{"mult"}},
	"new_wave":     {kind: kindPerCard, params: []string{"points", "mult"}},
	"rearguard":    {kind: kindPerCard, params: []string{"mult"}},
	"hidden":       {kind: kindBench, params: []string{"mult"}},
	"storehouse":   {kind: kindBench, params: []string{"points"}},
	"treasure":     {kind: kindOps, params: []string{"cap"}},
	"piggy":        {kind: kindWhole, params: []string{"per_gold", "mult"}},
	"thrifty":      {kind: kindOps, params: []string{"gold"}},
	"accountant":   {kind: kindWhole, params: []string{"per_sell"}, growth: true},
	"grudge":       {kind: kindWhole, params: []string{"per_loss", "cap"}, growth: true},
	"momentum":     {kind: kindWhole, params: []string{"mult"}},
	"last_stand":   {kind: kindWhole, params: []string{"hp_below", "factor"}, times: true},
	"tailwind":     {kind: kindWhole, params: []string{"per_level"}, times: true},
	"literary_pen": {kind: kindWhole, params: []string{"factor"}, times: true},
	"market_pass":  {kind: kindOps, params: []string{"free"}},
	"academy":      {kind: kindOps, params: []string{"xp"}},
	"expansion":    {kind: kindOps, params: []string{"slots"}},
	"tempering":    {kind: kindOps, params: []string{"points"}},
	"cinnabar":     {kind: kindPerCard, params: []string{"mult"}},
	"pine_ink":     {kind: kindPerCard, params: []string{"points"}},
	"shrine":       {kind: kindFactor, params: []string{"factor"}},
	"night_parade": {kind: kindWhole, params: []string{"min_cards", "factor"}, times: true},
	"one_color":    {kind: kindWhole, params: []string{"min_cards", "factor"}, times: true},
	"yin_yang":     {kind: kindWhole, params: []string{"min_each", "factor"}, times: true},
	"four_symbols": {kind: kindWhole, params: []string{"mult"}},
	"vermilion":    {kind: kindOps},
	"mirror":       {kind: kindCopy},
	"rubbing":      {kind: kindCopy},
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
	"quake":         {"bench"},
	"solar_eclipse": {"factor"},
	"ghost_rout":    {"factor"},
	"red_moon":      {"factor"},
	"ink_rain":      {"factor"},
	"famine":        {"income"},
	"market_closed": {"refresh"},
	"heavy_tax":     {},
	"shortage":      {"shop"},
	"narrow_path":   {"slots"},
	"discord":       {"at"},
	"upheaval":      {"factor"},
}

// familyParams: family key → required params.
var familyParams = map[string][]string{
	"merchant": {"gold"},                         // 开局 +金
	"scholar":  {"relics"},                       // 开局获得 N 个最低稀有度法宝
	"general":  {"level"},                        // 开局等级
	"artisan":  {"star"},                         // 开局第一张卡的星级
	"noble":    {"extra_offer"},                  // 宝库多 N 个选项
	"poor":     {"gold_per_loss"},                // 每输一回合额外 +金
	"miser":    {"interest_cap", "income_delta"}, // 利息上限 +N，基础收入 +delta
	"hermit":   {"relic_slots", "board_slots"},   // 法宝栏 +N，出阵格 +N（可为负）
}
