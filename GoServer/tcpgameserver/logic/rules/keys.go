// 配置表 key 的具名常量（data/config/tables/chess_*.yaml）。
// 表里新增被代码引用的 key 时在这里加常量；改名或删除表项时 keys_test.go 会失败。

package rules

// 代码里分支判断（switch / ==）一律用这些常量，不直接写字符串字面量。
// 每个常量的注释取自配置表中该项的名称与描述。

// 法宝（chess_relics.yaml 的 key）
const (
	RelicFurnace      = "furnace"       // 熔炉：每合成一次，永久 +2 气势
	RelicHeirloom     = "heirloom"      // 传家宝：2★ 卡结算时 +2 气势
	RelicMasterSeal   = "master_seal"   // 宗师印：出阵中每有一张 3★ 卡，气势 ×1.3
	RelicGraduation   = "graduation"    // 出师礼：每合成一次 +1 金
	RelicFellowship   = "fellowship"    // 同门帖：每个生效中的羁绊 +1.5 气势
	RelicSectToken    = "sect_token"    // 门派令：选定行当计数 +1；该行当生效时再 +3 气势
	RelicEclectic     = "eclectic"      // 杂家：出阵中至少有 6 个不同的系，气势 ×2
	RelicSecretArt    = "secret_art"    // 家传绝学：选定行当的卡在逐卡阶段再结算一次
	RelicLoneWolf     = "lone_wolf"     // 独行侠：出阵不超过 3 张，气势 ×2.5
	RelicLean         = "lean"          // 精兵简政：每个空着的出阵格 +3 气势
	RelicVanguard     = "vanguard"      // 先锋旗：最左边那张卡自己的战力 ×3
	RelicMultitude    = "multitude"     // 人多势众：每张出阵卡 +1 气势
	RelicNewWave      = "new_wave"      // 后浪：1★ 卡结算时 +4 战力
	RelicRearguard    = "rearguard"     // 殿后：最右边那张卡结算时 +5 气势
	RelicHidden       = "hidden"        // 藏锋：备战席每有一张 2★ 及以上的卡 +6 气势
	RelicStorehouse   = "storehouse"    // 库房：备战席每有一张卡 +15 战力
	RelicTreasure     = "treasure"      // 聚宝盆：利息上限 +3
	RelicPiggy        = "piggy"         // 存钱罐：结算时每持有 10 金 +1 气势
	RelicThrifty      = "thrifty"       // 精打细算：本回合没有刷新过（含免费刷新），结算时 +2 金
	RelicAccountant   = "accountant"    // 账房：每卖出一张 2★ 及以上的卡或一个法宝，永久 +2 气势
	RelicGrudge       = "grudge"        // 卧薪尝胆：每输一回合，永久 +1.5 气势，上限 +25
	RelicMomentum     = "momentum"      // 乘胜追击：进入本回合时每 1 场连胜 +3 气势
	RelicLastStand    = "last_stand"    // 背水一战：血量低于 30 时，气势 ×1.75
	RelicTailwind     = "tailwind"      // 乘风：气势 ×(1 + 0.15 × 等级)
	RelicLiteraryPen  = "literary_pen"  // 文曲笔：文墨羁绊生效时，气势 ×2
	RelicMirror       = "mirror"        // 铜镜：复制右边相邻法宝的效果（右边没有法宝时无效）
	RelicRubbing      = "rubbing"       // 拓本：复制最左边法宝的效果（自己在最左时无效）
	RelicMarketPass   = "market_pass"   // 市集牌：每回合第 1 次刷新免费
	RelicAcademy      = "academy"       // 书院荐：每回合 +2 经验（第 2 回合起）
	RelicExpansion    = "expansion"     // 扩建令：出阵格 +1
	RelicTempering    = "tempering"     // 百炼：合成出的卡永久 +3 战力
	RelicCinnabar     = "cinnabar"      // 朱砂：每张红色出阵卡结算时 +3 气势
	RelicPineInk      = "pine_ink"      // 松烟墨：每张黑色出阵卡结算时 +15 战力
	RelicShrine       = "shrine"        // 神龛：神卡在逐卡阶段战力 ×2
	RelicNightParade  = "night_parade"  // 百鬼夜行：出阵中鬼卡至少 4 张时，气势 ×2
	RelicOneColor     = "one_color"     // 一色：出阵卡全部同色（至少 3 张），气势 ×3
	RelicYinYang      = "yin_yang"      // 阴阳鱼：红黑张数相等（各至少 2 张），气势 ×2
	RelicFourSymbols  = "four_symbols"  // 四象：出阵中红神、黑神、红鬼、黑鬼齐全，+20 气势
	RelicVermilion    = "vermilion"     // 赤化：合成出的卡总是红色
	RelicDemonMirror  = "demon_mirror"  // 照妖镜：对手每张鬼卡在其逐卡阶段战力 ×0.7
	RelicGodBreaker   = "god_breaker"   // 破神符：对手每张神卡在其逐卡阶段战力 ×0.7
	RelicColorCounter = "color_counter" // 克色：对手出阵中较多的那种颜色，自己该颜色的出阵卡每张 +4 气势
	RelicBehead       = "behead"        // 斩首：对手最左边那张卡在逐卡阶段战力归零
	RelicUsurp        = "usurp"         // 反客为主：对手每有一个花色法宝，+5 气势
	RelicDefyHeaven   = "defy_heaven"   // 逆天：每有一条生效中的天象，气势 ×1.5
)

// 天象（chess_omens.yaml 的 key）
const (
	OmenQuake        = "quake"         // 地动（点数 1）：备战席格数 −2（放不下的卡自动卖出）
	OmenSolarEclipse = "solar_eclipse" // 天狗食日（点数 2）：神卡战力 ×0.4
	OmenGhostRout    = "ghost_rout"    // 百鬼退散（点数 3）：鬼卡战力 ×0.4
	OmenRedMoon      = "red_moon"      // 赤月（点数 4）：红色卡战力 ×0.6
	OmenInkRain      = "ink_rain"      // 墨雨（点数 5）：黑色卡战力 ×0.6
	OmenFamine       = "famine"        // 荒年（点数 6）：每回合基础收入 −2
	OmenMarketClosed = "market_closed" // 闭市（点数 7）：刷新费用 +1
	OmenHeavyTax     = "heavy_tax"     // 苛税（点数 8）：不产生利息
	OmenShortage     = "shortage"      // 断货（点数 9）：商店卡位 −1
	OmenNarrowPath   = "narrow_path"   // 窄道（点数 10）：出阵格 −1
	OmenDiscord      = "discord"       // 失和（点数 11）：所有羁绊生效所需数量 +1
	OmenUpheaval     = "upheaval"      // 天地异变（点数 12）：出阵中张数最多的花色（至少 2 张；并列时都算）战力 ×0.5
)

// 家世（chess_families.yaml 的 key）
const (
	FamilyMerchant = "merchant" // 商贾世家：开局 +6 金
	FamilyScholar  = "scholar"  // 书香门第：开局获得 1 个随机普通法宝
	FamilyGeneral  = "general"  // 将门之后：开局直接 2 级
	FamilyArtisan  = "artisan"  // 匠人之家：开局发的第一张卡直接是 2★
	FamilyNoble    = "noble"    // 名门望族：宝库改为四选一
	FamilyPoor     = "poor"     // 寒门：每输一回合额外 +1 金
	FamilyMiser    = "miser"    // 守财奴：利息上限 +3
	FamilyHermit   = "hermit"   // 方外之人：法宝栏 +1，出阵格 −1
)

// 系（chess_lines.yaml 的 key）
const (
	LineCarpenter = "carpenter" // 木灵 · 木魅（1 费，鬼）
	LineSmith     = "smith"     // 炉火 · 炉精（1 费，神）
	LineFarmer    = "farmer"    // 社稷 · 稻草精（1 费，鬼）
	LineTaoist    = "taoist"    // 玄门 · 黄巾力士（1 费，神）
	LineMonk      = "monk"      // 佛门 · 木鱼沙弥（1 费，神）
	LineScholar   = "scholar"   // 文昌 · 书蠹（1 费，鬼）
	LineDoctor    = "doctor"    // 杏林 · 参娃（2 费，神）
	LineArtist    = "artist"    // 丹青 · 墨灵（2 费，鬼）
	LineDesigner  = "designer"  // 造化 · 泥偶（2 费，鬼）
	LineTrader    = "trader"    // 财神 · 金蟾（2 费，神）
	LineSoldier   = "soldier"   // 战魂 · 阴兵（3 费，鬼）
	LineClerk     = "clerk"     // 阴司 · 鬼差（3 费，神）
)

// 行当羁绊（chess_bonds.yaml trades 的 key）
const (
	TradeCraft  = "craft"  // 天工：加气势
	TradeLetter = "letter" // 文墨：加气势
	TradeMonk   = "monk"   // 方外：加战力
	TradeCourt  = "court"  // 敕令：乘气势
)

// 出身羁绊（chess_bonds.yaml origins 的 key）
const (
	OriginRural  = "rural"  // 山野：加战力
	OriginMarket = "market" // 市井：加气势
	OriginHall   = "hall"   // 庙堂：乘气势
)

// 3★ 传奇技能（chess_lines.yaml legend.key）
const (
	LegendDamageTaken     = "damage_taken"       // 参娃：受到的伤害 ×0.7（向上取整）
	LegendExtraSlots      = "extra_slots"        // 鬼差：出阵格 +1（下回合起生效）
	LegendGrowOnMerge     = "grow_on_merge"      // 炉精：每次合成永久 +6 战力（每回合最多 3 次）
	LegendInterest        = "interest"           // 金蟾：利息改为每 10 金 +2，上限 +10
	LegendMultPerBond     = "mult_per_bond"      // 墨灵：每个生效中的羁绊 +2 气势
	LegendMultPerHighStar = "mult_per_high_star" // 木魅：出阵中每张 2★ 及以上的卡 +1 气势
	LegendMultPerLevel    = "mult_per_level"     // 书蠹：+气势，数值等于等级 × 2
	LegendMultPerLine     = "mult_per_line"      // 泥偶：出阵中每个不同的系 +1 气势
	LegendMultPerRelic    = "mult_per_relic"     // 木鱼沙弥：每持有一个法宝 +1 气势
	LegendOwnOriginCount  = "own_origin_count"   // 稻草精：山野羁绊计数 +1
	LegendRetriggerLeft   = "retrigger_left"     // 黄巾力士：最左边 1 个法宝额外触发 1 次
	LegendTimesMult       = "times_mult"         // 阴兵：气势 ×2
)

// 羁绊档位数值（chess_bonds.yaml tiers[].values 的 key）
const (
	BondPointsPerCard = "points_per_card" // 每张出阵卡 +点
	BondAddMult       = "add_mult"        // +气势（倍率）
	BondTimesMult     = "times_mult"      // 气势 ×倍数
)

// 卡牌花色在配置中的写法（对应 Color 枚举 Red / Black）
const (
	ColorNameRed   = "red"   // 赤（神）
	ColorNameBlack = "black" // 墨（鬼）
)
