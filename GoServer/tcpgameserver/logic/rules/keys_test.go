package rules

import "testing"

// 每个具名常量都必须仍是配置表里的 key（改名或删除表项时这里会失败）。
func TestKeyConstantsMatchTables(t *testing.T) {
	tt := loadTestTables(t)
	has := func(what string, keys []string, want ...string) {
		t.Helper()
		set := make(map[string]bool, len(keys))
		for _, k := range keys {
			set[k] = true
		}
		for _, w := range want {
			if !set[w] {
				t.Errorf("%s key %q not found in tables", what, w)
			}
		}
	}

	var relics, omens, families, lines, legends, trades, origins []string
	for _, r := range tt.Rel.Relics {
		relics = append(relics, r.Key)
	}
	for _, o := range tt.O.Omens {
		omens = append(omens, o.Key)
	}
	for _, f := range tt.F.Families {
		families = append(families, f.Key)
	}
	for _, l := range tt.L.Lines {
		lines = append(lines, l.Key)
		if l.Legend.Key != "" {
			legends = append(legends, l.Legend.Key)
		}
	}
	for _, b := range tt.B.Trades {
		trades = append(trades, b.Key)
	}
	for _, b := range tt.B.Origins {
		origins = append(origins, b.Key)
	}

	has("relic", relics, RelicFurnace, RelicHeirloom, RelicMasterSeal, RelicGraduation, RelicFellowship, RelicSectToken, RelicEclectic, RelicSecretArt, RelicLoneWolf, RelicLean, RelicVanguard, RelicMultitude, RelicNewWave, RelicRearguard, RelicHidden, RelicStorehouse, RelicTreasure, RelicPiggy, RelicThrifty, RelicAccountant, RelicGrudge, RelicMomentum, RelicLastStand, RelicTailwind, RelicLiteraryPen, RelicMirror, RelicRubbing, RelicMarketPass, RelicAcademy, RelicExpansion, RelicTempering, RelicCinnabar, RelicPineInk, RelicShrine, RelicNightParade, RelicOneColor, RelicYinYang, RelicFourSymbols, RelicVermilion, RelicDemonMirror, RelicGodBreaker, RelicColorCounter, RelicBehead, RelicUsurp, RelicDefyHeaven)
	has("omen", omens, OmenQuake, OmenSolarEclipse, OmenGhostRout, OmenRedMoon, OmenInkRain, OmenFamine, OmenMarketClosed, OmenHeavyTax, OmenShortage, OmenNarrowPath, OmenDiscord, OmenUpheaval)
	has("family", families, FamilyMerchant, FamilyScholar, FamilyGeneral, FamilyArtisan, FamilyNoble, FamilyPoor, FamilyMiser, FamilyHermit)
	has("line", lines, LineCarpenter, LineSmith, LineFarmer, LineTaoist, LineMonk, LineScholar, LineDoctor, LineArtist, LineDesigner, LineTrader, LineSoldier, LineClerk)
	has("trade", trades, TradeCraft, TradeLetter, TradeMonk, TradeCourt)
	has("origin", origins, OriginRural, OriginMarket, OriginHall)
	has("legend", legends, LegendDamageTaken, LegendExtraSlots, LegendGrowOnMerge, LegendInterest, LegendMultPerBond, LegendMultPerHighStar, LegendMultPerLevel, LegendMultPerLine, LegendMultPerRelic, LegendOwnOriginCount, LegendRetriggerLeft, LegendTimesMult)
}
