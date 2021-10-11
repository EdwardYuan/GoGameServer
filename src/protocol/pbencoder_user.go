package message

import (
	"hope_server/game/gmevt"

	"pb/shared"
)

func (EncGame) UserData(e *gmevt.UserData) *shared.UserData {
	if e == nil {
		return nil
	}
	return &shared.UserData{
		Player:              encGame.Player(e.Player),
		Quests:              encGame.Quests(e.Quests),
		Npcs:                encGame.NPCs(e.NPCs),
		Friends:             encGame.Friends(e.Friends),
		Blacklist:           encGame.Blacklists(e.Blacklists),
		Mounts:              encGame.Mounts(e.Mounts),
		Partners:            encGame.Partners(e.Partners),
		Talents:             encGame.Talents(e.Talents),
		Skills:              encGame.Skills(e.Skills),
		SkillChooses:        encGame.SkillChooses(e.SkillChooses),
		Fames:               encGame.Fames(e.Fames),
		MapBuffs:            encGame.MapBuffs(e.MapBuffs),
		Bag:                 encGame.Bag(e.Bag),
		Achievements:        encGame.PlayerAchievements(e.Achievements),
		FormulaIds:          encGame.FormulaIDs(e.FormulaIDs),
		Runes:               encGame.Runes(e.Runes),
		RunePages:           encGame.RunePages(e.RunePages),
		Signs:               encGame.Signs(e.Signs),
		Equips:              encGame.Equipments(e.Equips),
		BagExtends:          encGame.BagExtends(e.BagExtends),
		Tools:               encGame.Tools(e.Tools),
		KvStores:            encGame.KVStores(e.KVStores),
		Coupons:             encGame.Coupons(e.Coupons),
		Payments:            encGame.Payments(e.Payments),
		Jewels:              encGame.Jewels(e.Jewels),
		FinishFormationInfo: encGame.FinishFormationInfo(e.FinishFormationInfo),
		Grids:               encGame.equipGrids(e.EquipGrids),
		LifeSkills:          encGame.LifeSkills(e.LifeSkills),
		ActivatedMounts:     encGame.ActivatedMounts(e.ActivatedMounts), //encGame.Mounts(e.ActivatedMounts),
		ActiveRewards:       encGame.ActiveRewards(e.ActiveRewards),

		// Version: &shared.UserVersion{
		// 	Player:              encGame.PlayerVer(e.Player),
		// 	Quests:              encGame.PlayerQuestsVer(e.Quests),
		// 	Friends:             encGame.PlayerFriendsVer(e.Friends),
		// 	Blacklist:           encGame.PlayerBlacklistVer(e.Blacklists),
		// 	Mounts:              encGame.PlayerMountsVer(e.Mounts),
		// 	Partners:            encGame.PlayerPartnersVer(e.Partners),
		// 	Talents:             encGame.PlayerTalentsVer(e.Talents),
		// 	Skills:              encGame.PlayerSkillsVer(e.Skills),
		// 	SkillChooses:        encGame.PlayerSkillChoosesVer(e.SkillChooses),
		// 	Fames:               encGame.PlayerFamesVer(e.Fames),
		// 	MapBuffs:            encGame.PlayerMapBuffVer(e.MapBuffs),
		// 	Bag:                 encGame.PlayerBagVer(e.Bag),
		// 	Achievements:        encGame.PlayerAchievementVer(e.Achievements),
		// 	FormulaIds:          encGame.PlayerFormulaIDVer(e.FormulaIDs),
		// 	Runes:               encGame.PlayerRuneVer(e.Runes),
		// 	RunePages:           encGame.PlayerRunePageVer(e.RunePages),
		// 	Signs:               encGame.PlayerSignVer(e.Signs),
		// 	Equips:              encGame.PlayerEquipsVer(e.Equips),
		// 	BagExtends:          encGame.PlayerBagExtendsVer(e.BagExtends),
		// 	Tools:               encGame.PlayerToolsVer(e.Tools),
		// 	KvStores:            encGame.PlayerKVStoresVer(e.KVStores),
		// 	Coupons:             encGame.PlayerCouponsVer(e.Coupons),
		// 	Payments:            encGame.PlayerPaymentsVer(e.Payments),
		// 	Jewels:              encGame.PlayerJewelsVer(e.Jewels),
		// 	FinishFormationInfo: encGame.PlayerFinishFormationInfoVer(e.FinishFormationInfo),
		// 	EquipGrid:           encGame.equipGridVer(e.EquipGrids),
		// 	LifeSkills:          encGame.PlayerLifeSkillsVer(e.LifeSkills),
		// 	ActivatedMounts:     encGame.PlayerMountsVer(e.ActivatedMounts),
		// 	ActiveRewards:       encGame.PlayerActiveRewardsVer(e.ActiveRewards),
		// },
	}
}

func (EncGame) Operation(e *gmevt.Operation) *shared.Operation {
	if e == nil {
		return nil
	}
	var op shared.Operation_Opt
	if *e.Op == gmevt.OpInsert {
		op = shared.Operation_insert
	} else if *e.Op == gmevt.OpUpdate {
		op = shared.Operation_update
	} else if *e.Op == gmevt.OpDelete {
		op = shared.Operation_delete
	}

	// mark error if more than one change being put into Operation

	return &shared.Operation{
		Opt: op,
		// Version:         encGame.Version(e.Version),
		Player:          encGame.Player(e.Player), //will be nil if e.Player is nil
		Quest:           encGame.Quest(e.Quest),
		Npc:             encGame.NPC(e.NPC),
		Item:            encGame.Item(e.Item),
		Friend:          encGame.Friend(e.Friend),
		Blacklist:       encGame.Blacklist(e.Blacklist),
		Mount:           encGame.Mount(e.Mount),
		Partner:         encGame.Partner(e.Partner),
		Talent:          encGame.Talent(e.Talent),
		Skill:           encGame.Skill(e.Skill),
		SkillChoose:     encGame.SkillChoose(e.SkillChoose),
		Fame:            encGame.Fame(e.Fame),
		MapBuff:         encGame.MapBuff(e.MapBuff),
		BagCell:         encGame.ItemCell(e.BagCell),
		Achievement:     encGame.Achievement(e.Achievement),
		FormulaId:       encGame.FormulaID(e.FormulaID),
		Rune:            encGame.Rune(e.Rune),
		RunePage:        encGame.RunePage(e.RunePage),
		Sign:            encGame.Sign(e.Sign),
		Equip:           encGame.Equipment(e.Equip),
		Tool:            encGame.Tool(e.Tool),
		BagExtend:       encGame.BagExtend(e.BagExtend),
		Coupon:          encGame.Coupon(e.Coupon),
		Payment:         encGame.Payment(e.Payment),
		Jewel:           encGame.Jewel(e.Jewel),
		FinishFormation: encGame.FinishFormation(e.FinishFormation),
		Grid:            encGame.equipGrid(e.EquipGrid),
		LifeSkill:       encGame.LifeSkill(e.LifeSkill),
		ActivatedMount:  encGame.ActivatedMount(e.ActivatedMount), //encGame.Mount(e.ActivatedMount),
		ActiveReward:    encGame.ActiveReward(e.ActiveReward),
	}
}

func (EncGame) equipGrids(e *gmevt.PlayerEquipGrids) []*shared.EquipGrid {
	if e == nil {
		return nil
	}
	rv := []*shared.EquipGrid{}
	for _, grid := range e.EquipGrids {
		rv = append(rv, encGame.equipGrid(grid))
	}
	return rv
}

func (EncGame) equipGrid(e *gmevt.EquipGrid) *shared.EquipGrid {
	if e == nil {
		return nil
	}
	return &shared.EquipGrid{
		Location: int32(e.Location),
		Level:    int32(e.Level),
		//FailedTimes: proto.Int(e.FailedTimes),
	}
}

// func (EncGame) equipGridVer(e *gmevt.PlayerEquipGrids) *shared.Version {
// 	if e == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(e.Rev)
// 	return encGame.Version(&v)
// }
