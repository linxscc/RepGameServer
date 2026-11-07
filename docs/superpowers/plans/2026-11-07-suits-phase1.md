# 花色第一阶段（花色 + 卡池 + 合成 + 模拟）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 每张卡有花色（系决定神/鬼，单张卡决定红/黑），卡池翻倍并按红黑分开计数，合成取多数颜色，客户端卡面显示花色，并用模拟器回答设计文档 2.4 的三个问题。

**Architecture:** 规则引擎（`rules`）给 `Card` 加颜色和"含红色张数"，卡池在原有总数 `Pool` 之外加 `PoolRed`，商店卡位加平行的 `ShopColor`。协议新增两个枚举，视图层填字段，Unity 卡面读取。模拟实验写在 `GoServer/tools/balance`（本地工具，按用户规则不提交）。

**Tech Stack:** Go（tcpgameserver，yaml 配置，protobuf），Unity 6 C#（UGUI + TMP），protoc 3.29.3。

**Spec:** `docs/superpowers/specs/2026-11-07-suits-and-omens-design.md`（第 2 节与 6.1）

## Global Constraints

- 服务端提交并推送到 `RepServer` 分支（推送即部署）；客户端提交并推送到 RepGame `main`。不开 PR，提交信息不加任何 Co-Authored-By / Claude 署名行。
- `GoServer/tools/` 不提交。
- 卡池翻倍：1 费 26、2 费 22、3 费 18，红黑各半。
- 合成：同系同星 3 张，不要求同色；结果颜色 = 3 张中较多的颜色。
- 神鬼分配：神 = 黄巾力士 taoist、木鱼沙弥 monk、炉精 smith、参娃 doctor、金蟾（2 费，查 key）、鬼差（3 费，查 key）；其余为鬼。
- 红色字用 `Theme.Red`，黑色字用 `Theme.Ink`。
- 协议重新生成：`scripts/gen-proto.sh`，`PROTOC=C:\Users\z1748\AppData\Local\Temp\claude\9bae702da576987b\820f3fc9-b65d-4c1d-a089-c0c57ea6edd8\scratchpad\pbtools\pkgs\google.protobuf.tools\3.29.3\tools\windows_x64\protoc.exe`，PATH 包含 `/c/Users/z1748/go/bin`；同时生成 `pb/game.pb.go` 与客户端 `Game.cs`。
- 客户端编译检查：`cd C:\Users\z1748\AppData\Local\Temp\claude\9bae702da576987b\da0054fd-bd12-470e-9c93-aaec14867f84\scratchpad\clientcheck && dotnet build ClientCheck.csproj -nologo`，期望 `0 个错误`。

## Review Focus

1. 卖出红黑混合的合成卡：必须按卡内实际红/黑张数归还卡池，任何时候 `PoolRed[l]` 不为负、不超过 `Pool[l]`（测试在 Task 2）。
2. 某系的一种颜色抽光后，商店只抽另一种颜色，不能抽出负数（Task 2）。
3. 锁定商店跨回合时颜色保持不变（Task 2）。
4. 家世"匠人之家"开局 2★ 卡：颜色与 `Red` 一致（3 张拷贝各自抽色）（Task 2）。
5. 断线重连收到的状态里，出阵/备战/商店卡的颜色与断线前一致（Task 3 视图测试）。

---

### Task 1: 配置——神鬼、卡池翻倍、同色合成开关

**Files:**
- Modify: `GoServer/tcpgameserver/rules/config.go`（`LineDef`、merge 配置结构、校验）
- Modify: `GoServer/tcpgameserver/config/tables/chess_lines.yaml`（每个系加 `kind`，`pool` 翻倍，表头注释加 `kind` 说明）
- Modify: `GoServer/tcpgameserver/config/tables/chess_rules.yaml`（`merge.same_color: false`，注释"true 时只有同色才能合成（模拟对照用）"）
- Test: `GoServer/tcpgameserver/rules/rules_test.go`

**Interfaces:**
- Produces: `LineDef.Kind string`（`yaml:"kind"`，取值 `"god"` / `"ghost"`）；常量 `KindGod = "god"`、`KindGhost = "ghost"`；merge 配置新增 `SameColor bool`（`yaml:"same_color"`）。

- [ ] **Step 1: 写失败测试 `TestLinesHaveKindAndEvenPool`**

```go
func TestLinesHaveKindAndEvenPool(t *testing.T) {
	cfg := loadTestConfig(t)
	gods := map[int]int{}
	for _, l := range cfg.Lines {
		if l.Kind != KindGod && l.Kind != KindGhost {
			t.Fatalf("%s kind %q", l.Key, l.Kind)
		}
		if l.Pool%2 != 0 {
			t.Fatalf("%s pool %d not even", l.Key, l.Pool)
		}
		want := map[int]int{1: 26, 2: 22, 3: 18}[l.Cost]
		if l.Pool != want {
			t.Fatalf("%s pool %d want %d", l.Key, l.Pool, want)
		}
		if l.Kind == KindGod {
			gods[l.Cost]++
		}
	}
	if gods[1] != 3 || gods[2] != 2 || gods[3] != 1 {
		t.Fatalf("gods per cost %v", gods)
	}
	if cfg.Rules.Merge.SameColor {
		t.Fatal("same_color must default to false")
	}
}
```

- [ ] **Step 2: 运行确认失败** — `cd GoServer && go test ./tcpgameserver/rules -run TestLinesHaveKindAndEvenPool`，期望编译失败（`KindGod` 未定义）。
- [ ] **Step 3: 实现** — 加字段与常量；校验：`kind` 只能是两种之一、`pool` 为正偶数（错误信息格式沿用 `chess_lines %s ...`）；改 yaml。
- [ ] **Step 4: 运行通过** — `go test ./tcpgameserver/...`，全部 PASS（其他测试若依赖旧卡池数量，按新数量改断言）。
- [ ] **Step 5: 提交** — `git commit -m "rules: line kind (god/ghost), doubled pools, merge.same_color switch"`

### Task 2: 规则——卡的颜色、红黑卡池、合成取多数色

**Files:**
- Modify: `rules/state.go`（`Card`、`Player`、`Match` 字段；`CountCopies`）
- Modify: `rules/match.go`（`NewMatch` 初始化 `PoolRed`；`dealStart`；`CheckPool`）
- Modify: `rules/actions.go`（`drawShopCard`、`returnShop`、`rollShop`、`BuyCard`、`checkMerges`、`merge`、`disposeCard`）
- Modify: `rules/helpers_test.go`（`forceShop` 按颜色；新增 `forceShopColored`、`coloredCard`）
- Test: `rules/rules_test.go`

**Interfaces:**
- Produces:
  - `type Color int`；`const ( Red Color = 0; Black Color = 1 )`
  - `Card.Color Color`、`Card.Red int`（卡内包含的红色 1★ 拷贝数；1★ 红卡为 1，黑卡为 0）
  - `Match.PoolRed []int`（各系剩余红色张数；黑色 = `Pool[l] - PoolRed[l]`）
  - `Player.ShopColor []Color`（与 `Shop` 等长；`Shop[i] < 0` 时该位颜色无意义）
  - `drawShopCard(p *Player) (line int, c Color)`
  - `func (m *Match) CountCopiesColor(p *Player, line, star int, c Color) int`
- 测试辅助：`forceShopColored(t, m, p, []string{"soldier"}, []Color{Red})`；`forceShop` 保持签名，颜色一律取 `Red`（红抽光则 `Black`）。

规则要点（实现者照此写，不另行设计）：
- 抽卡：先按现有逻辑选系（权重仍用 `Pool[l]`），再按该系剩余红/黑张数加权选颜色。
- 合成：`SameColor == false` 时分组只看系和星；`true` 时同系同星同色才计数。参与合成的 3 张中多数颜色为结果颜色；`Red` = 3 张的 `Red` 之和。合成时从出阵/备战取卡的顺序不变。
- 卖出 / 回收：`Pool[l] += copiesOf(star)`，`PoolRed[l] += c.Red`。
- 开局发卡：每张拷贝各自从卡池按红黑加权抽一次，`Red` 为抽到的红色数，`Color` 为多数色（1★ 就是那一张的颜色）。

- [ ] **Step 1: 写失败测试**

```go
func TestShopSlotsCarryColor(t *testing.T)            // 刷新 50 次：len(ShopColor)==len(Shop)，每次 CheckPool()
func TestMergeTakesMajorityColor(t *testing.T)        // 备战放红、红，商店 forceShopColored 黑 soldier，买入 → 合成卡 Color==Red、Red==2、Star==2
func TestSellMergedReturnsExactColors(t *testing.T)   // 接上：卖出 → Pool[soldier] +3、PoolRed[soldier] +2、CheckPool()
func TestSameColorMergeRule(t *testing.T)             // cfg.Rules.Merge.SameColor=true：红红黑不合成；再买一张红 → 合成，Red==3
func TestShopDrawsOtherColorWhenOneExhausted(t *testing.T) // 把某系 PoolRed 设 0：抽 200 次该系全是 Black，PoolRed 从不 <0
func TestLockedShopKeepsColors(t *testing.T)          // 锁定后进入下一回合，ShopColor 不变
func TestArtisanStartCardColorConsistent(t *testing.T)// 匠人之家：开局 2★ 卡 0<=Red<=3，Color 与 Red 的多数一致，CheckPool()
```

`CheckPool` 同时校验：各系 `PoolRed + 商店中红色 + 持有卡 Red 之和 == def.Pool/2`。

- [ ] **Step 2: 运行确认失败** — `go test ./tcpgameserver/rules`，期望编译失败（`Red`、`ShopColor` 未定义）。
- [ ] **Step 3: 实现** 上述字段与规则。
- [ ] **Step 4: 运行通过** — `go test ./tcpgameserver/...`（含 `chessbot` 的 `CheckPool` 长局测试）全部 PASS；`go vet ./tcpgameserver/...` 无输出。
- [ ] **Step 5: 提交** — `git commit -m "rules: card colors, red/black pool, majority-color merge"`

### Task 3: 协议与视图

**Files:**
- Modify: `proto/repgame/v1/game.proto`
- Regenerate: `GoServer/tcpgameserver/pb/game.pb.go`、客户端 `Assets/Scripts/Protocol/Generated/Game.cs`
- Modify: `GoServer/tcpgameserver/logic/chess_views.go`（`chessCards`、商店卡位、`LineInfo`）
- Test: `GoServer/tcpgameserver/logic/chess_views_test.go`（无则新建）

**Interfaces:**
- Produces（proto）：
  - `enum CardColor { CARD_COLOR_UNSPECIFIED = 0; CARD_COLOR_RED = 1; CARD_COLOR_BLACK = 2; }`
  - `enum CardKind { CARD_KIND_UNSPECIFIED = 0; CARD_KIND_GOD = 1; CARD_KIND_GHOST = 2; }`
  - `ChessCard.color = 8`（CardColor）、`ShopSlot.color = 4`（CardColor）、`LineInfo.kind = 9`（CardKind）
  - C# 侧：`ChessCard.Color`、`ShopSlot.Color`、`LineInfo.Kind`

- [ ] **Step 1: 写失败测试 `TestViewsCarryColorAndKind`** — 建对局，强制一张红 soldier 在出阵、商店一张黑卡；断言视图里 `ChessCard.Color == CARD_COLOR_RED`、对应商店位 `Color == CARD_COLOR_BLACK`、MatchConfig 每条 `LineInfo.Kind` 非 UNSPECIFIED 且与配置一致。
- [ ] **Step 2: 运行确认失败。**
- [ ] **Step 3: 改 proto、重新生成、填视图字段。**
- [ ] **Step 4: 运行通过** — `go build ./... && go test ./tcpgameserver/...`。
- [ ] **Step 5: 提交服务端**（proto、pb、views、test）`git commit -m "proto: card color and kind"`。生成的 `Game.cs` 留到 Task 4 随客户端一起提交。

### Task 4: 客户端卡面显示花色

**Files:**
- Modify: `RepGame/Assets/Scripts/Chess/UI/CardView.cs`（`Build` 画花色字；`Describe` 加"花色"行）
- Modify: `RepGame/Assets/Scripts/Chess/UI/ShopViews.cs:164`（预览卡带 `Color = slot.Color`）
- Commit 包含：`Assets/Scripts/Protocol/Generated/Game.cs`

**Interfaces:**
- Consumes: Task 3 的 `ChessCard.Color`、`ShopSlot.Color`、`LineInfo.Kind`。
- Produces: `static string CardView.SuitName(ChessCard c)` → `"红神" / "黑神" / "红鬼" / "黑鬼"`，未知返回 `""`。

- [ ] **Step 1: 实现** — 左下角、战力徽章上方画一个字（神或鬼），字号约 `w * 0.15f`，加粗，颜色红 `Theme.Red` / 黑 `Theme.Ink`，底下垫一个半透明宣纸圆（`Theme.WithAlpha(Theme.Paper, 0.9f)`），保证在任何卡图上都清楚。颜色或种类未知时不画。`Describe` 的副标题（sub）末尾追加 ` · 红神` 这样的花色。
- [ ] **Step 2: 编译检查** — 跑客户端编译命令，期望 `0 个错误`。
- [ ] **Step 3: 提交并推送客户端** — `git add -f` 不需要（只有 .cs）；`git commit -m "client: show card suit (red/black 神/鬼)"`；先不推送，等 Task 6 模拟通过后与服务端一起推送（见 Task 7）。

### Task 5: 机器人按颜色买卡

**Files:**
- Modify: `GoServer/tcpgameserver/chessbot/bot.go`
- Test: `GoServer/tcpgameserver/chessbot/bot_test.go`

**Interfaces:**
- Produces: `Archetype.Color string`（`""` 不看颜色，`"red"` / `"black"` 偏好该色）；导出 `func ColorPeak(m *rules.Match, p *rules.Player) int`（出阵中同色卡的最大张数，供模拟统计）。

行为：偏好某色时，商店里同系有两种颜色可选，只买偏好色；为了合成缺的那一张不论颜色照买（合成优先）；排阵时同等条件下优先上偏好色。`""` 时行为与现在完全一致。

- [ ] **Step 1: 写失败测试 `TestColorBotFieldsMostlyItsColor`** — 两个 bot（一个 `Color: "red"`，一个 `""`），固定 10 个种子各跑完整局；断言红色 bot 在第 15 回合结算时 `ColorPeak` 的平均值 ≥ 无偏好 bot 的平均值 + 1。
- [ ] **Step 2: 运行确认失败。**
- [ ] **Step 3: 实现。**
- [ ] **Step 4: 运行通过** — `go test ./tcpgameserver/...`。
- [ ] **Step 5: 提交** — `git commit -m "chessbot: color preference"`

### Task 6: 模拟实验（不提交）

**Files:**
- Modify: `GoServer/tools/balance/main.go`（新实验 `-exp suits`）

记录：每局每方的 3★ 张数、首个 3★ 出现回合（没有记 0，统计时排除）、第 10/15/20 回合结算时的 `ColorPeak`。

实验组（每组 `-n 2000`，随机 archetype 对战，与现有 `expMatrix` 相同的 bot 池）：

| 组 | 卡池 | 合成 |
|---|---|---|
| 旧 | 当前数量 ÷2（即改前的 13/11/9） | 不分色 |
| A | 翻倍 | 不分色（方案 A，默认） |
| B | 翻倍 | 同色才合成 |
| A-1.5 | 1.5 倍（20/16/14，取偶数） | 不分色 |

另加一组颜色流派：红偏好 bot vs 无偏好 bot，卡池翻倍、方案 A，报告双方 `ColorPeak` 均值和胜率。

输出表格：每方平均 3★ 张数、首个 3★ 平均回合、平均局长、`ColorPeak` 均值。

- [ ] **Step 1: 实现实验**（用 `loadCfg(mut)` 改卡池与 `SameColor`）。
- [ ] **Step 2: 运行** — `cd GoServer && go run ./tools/balance -exp suits -n 2000`，保存输出。
- [ ] **Step 3: 把结果表格和结论发给用户**，回答设计文档 2.4 的三个问题，并建议卡池是否下调（用户已允许下调，保持红黑各半）。按用户决定改 `chess_lines.yaml` 的 `pool`，重跑 Task 1 测试（若改了数量，同步修改 Task 1 测试里的期望值）。

### Task 7: 上线

- [ ] **Step 1:** `go build ./... && go vet ./tcpgameserver/... && go test ./tcpgameserver/...` 全部通过。
- [ ] **Step 2:** 推送服务端 `git push origin RepServer`（触发部署）。
- [ ] **Step 3:** 推送客户端 `git push origin main`。
- [ ] **Step 4:** 告诉用户可以试玩，以及没有在 Unity 里验证过的部分。
