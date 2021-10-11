//NOTE: the mechanism in this file is kind of a hack, please BE AWARE!
package message

import (
	"errors"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"hope_server/config"
	"hope_server/event"
	"hope_server/game/gmevt"
	"hope_server/intutil"
	"hope_server/kit/logz"

	"hope_server/common/user"
	"pb/action"
	"pb/base"
	"pb/chat"
	"pb/field"
	"pb/friend"
	"pb/kvstore"
	"pb/query"
	"pb/shared"
)

var ErrUnknown = errors.New("unknown error")

func GenerateEventProtobuf(seq int, e event.Event) (bs []byte, name string, headerLen, bodyLen int, err error) {
	if !config.FlagDev {
		defer func() {
			if r := recover(); r != nil {
				logz.Error("recover from proto event generate error", "event_name", event.Name(e),
					"recover_result", r, "debug_stack", string(debug.Stack()))
				switch r := r.(type) {
				case string:
					err = errors.New(r)
				case error:
					err = r
				default:
					err = ErrUnknown
				}
			}
		}()
	}

	out := MessageProtoFromEvent(seq, e)
	return out.Encode()
}

func MessageProtoFromEvent(seq int, e event.Event) *Message {
	var m proto.Message

	//NOTE: assume concrete event is a pointer
	t := reflect.ValueOf(e).Elem().Type()
	pkgSlice := strings.Split(t.PkgPath(), "/")
	pkgName := pkgSlice[len(pkgSlice)-1]
	eventName := pkgName + "." + t.Name()

	//println("GenerateEventProtobuf:", eventName)
	//	methodValue, ok := encMethods[eventName]
	//	if !ok {
	//		panic(: no registered method for event " + eventName)
	//	}

	methodValue, ok := encMethods[eventName]
	if ok {
		// 手动从event转到protobuf
		arg := reflect.ValueOf(e)
		rv := methodValue.Call([]reflect.Value{arg})
		m, ok = rv[0].Interface().(proto.Message)
		if !ok {
			panic("Cannot type assert return value to proto.Message")
		}
		if m == nil {
			panic("nil message probably should not reach here")
		}
	} else {
		// 直接编码
		m = e.(proto.Message)
	}
	return NewMessageProto(int32(seq), m)
}

type PBEncoder interface {
	PkgName() string
}

//type EncBattle struct{}
type EncGame struct{}

//func (e EncBattle) PkgName() string {
//	return "btevt"
//}
func (e EncGame) PkgName() string {
	return "gmevt"
}

var (
	//	encBattle EncBattle
	encGame EncGame
)

var encMethods map[string]reflect.Value = map[string]reflect.Value{}

func init() {
	//	registerEncoder(encBattle)
	registerEncoder(encGame)
}

func registerEncoder(enc PBEncoder) {
	v := reflect.ValueOf(enc)
	t := v.Type()
	n := t.NumMethod()
	for i := 0; i < n; i++ {
		method := t.Method(i)
		if method.Name == "PkgName" { // this is in interface{}
			continue
		}
		name := enc.PkgName() + "." + method.Name
		if _, ok := encMethods[name]; ok {
			panic("encode method " + name + " already registered")
		}
		encMethods[name] = v.Method(i)
	}
}

func (EncGame) Pong(e *gmevt.Pong) *shared.Pong {
	return new(shared.Pong)
}

func (EncGame) ServerMessage(e *gmevt.ServerMessage) *shared.ServerMessage {
	return &shared.ServerMessage{
		Content: e.Content,
		Code:    e.Code,
	}
}

func fieldPos(x int) *field.FieldPos {
	return &field.FieldPos{X: int32(x)}
}

func fieldEquips(equips []gmevt.Equip) []*field.Equip {
	rv := []*field.Equip{}
	for _, v := range equips {
		rv = append(rv, &field.Equip{
			Id: int32(v.ID),
			//RecastTimes: proto.Int(v.RecastTimes),
			FashionID: int32(v.FashionID),
		})
	}

	return rv
}

/*func teamEquips(equips []gmevt.Equip) []*team.Equip {
	rv := []*team.Equip{}
	for _, v := range equips {
		rv = append(rv, &team.Equip{
			Id:          proto.Int(v.ID),
			//RecastTimes: proto.Int(v.RecastTimes),
			FashionID:   proto.Int(v.FashionID),
		})
	}

	return rv
} Banson changed team proto */

func (EncGame) ResEnterField(e *gmevt.ResEnterField) *field.S2C_EnterField {
	if e == nil {
		return nil
	}
	return &field.S2C_EnterField{
		Pos:        fieldPos(e.Pos),
		FieldID:    int32(e.FieldID),
		FieldSubID: e.FieldSubID,
	}
}

func (EncGame) ResEnterHome(e *gmevt.ResEnterHome) *field.S2C_EnterHome {
	if e == nil {
		return nil
	}
	return &field.S2C_EnterHome{
		Field: fieldMap(e.Field, ""),
		Pos:   fieldPos(e.Pos),
	}
}

func (EncGame) SrvFieldUpdate(e *gmevt.SrvFieldUpdate) *field.S2C_Update {
	if e == nil {
		return nil
	}
	return &field.S2C_Update{
		Entities: fieldStates(e.EntityStates),
	}
}

func (EncGame) ResQueryUser(e *gmevt.ResQueryUser) *query.S2C_QueryUser {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryUser{
		UserData:       encGame.UserData(e.UserData),
		FieldStates:    fieldStates(e.EntityStates),
		TeamingState:   nil, //Banson changed for the new "team" function
		BattleId:       e.BattleID,
		BattleUrl:      e.BattleURL,
		BattleAuthCode: e.BattleAuthCode,
		HotDBMD5:       e.HotDBMD5,
	}
}

func fieldStates(states []gmevt.EntityState) []*field.Entity {
	rv := []*field.Entity{}
	for _, s := range states {
		rv = append(rv, entityState(s))
	}
	return rv
}

func entityState(es gmevt.EntityState) *field.Entity {
	if es.PlayerState != nil {
		return &field.Entity{State: &field.Entity_Player{playerState(es.PlayerState)}}
	}
	if es.NPCState != nil {
		return &field.Entity{State: &field.Entity_Npc{npcState(es.NPCState)}}
	}
	return nil
}

func playerState(s *gmevt.PlayerFieldState) *field.PlayerState {
	followers := make([]*field.Entity, len(s.Followers))
	for i, f := range s.Followers {
		followers[i] = entityState(f)
	}
	fashionIDs := []int32{}
	for _, id := range s.FashionIDs {
		fashionIDs = append(fashionIDs, int32(id))
	}
	return &field.PlayerState{
		UserBaseInfo: &base.UserBaseInfo{
			Uid:          int64(s.UID),
			Name:         s.UserBaseInfo.Name,
			Lv:           int32(s.Lv),
			RaceID:       int32(s.RaceID),
			ClassID:      int32(s.ClassID),
			SpecialistID: int32(s.SpecialistID),
			Gender:       int32(s.Gender),
			FaceID:       int32(s.FaceID),
			HairID:       int32(s.HairID),
			HairColorID:  int32(s.HairColorID),
			UidIM:        s.UIDIM,
			FashionIDs:   fashionIDs,
			MountID:      int32(s.MountID),
			BadgeID:      int32(s.BadgeID),
			GuildName:    s.GuildName,
		},
		Visibility: field.Visibility(s.Visibility),
		Pos:        fieldPos(s.Pos),
		Action:     field.FieldAction(s.Action.ActionType),
		ActionPar1: s.Action.Par1,
		Map:        fieldMap(s.FieldID, s.FieldSubID),
		Followers:  followers,
	}
}

func npcState(s *gmevt.NPCFieldState) *field.NPCState {
	return &field.NPCState{
		Uid:        s.NPCUID,
		Id:         int32(s.NPCID),
		Name:       s.Name,
		Pos:        fieldPos(s.Pos),
		Map:        fieldMap(s.FieldID, s.FieldSubID),
		Visibility: field.Visibility(s.Visibility),
	}
}

func (EncGame) ResQueryOtherUser(e *gmevt.ResQueryOtherUser) *query.S2C_QueryOtherUser {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryOtherUser{
		Ok:       e.Ok,
		Who:      int64(e.Who),
		UserData: encGame.UserData(e.UserData),
	}
}

func (EncGame) EnterBattle(e *gmevt.EnterBattle) *field.S2C_EnterBattle {
	if e == nil {
		return nil
	}
	return &field.S2C_EnterBattle{
		Id:          e.ID,
		Url:         e.URL,
		AuthCode:    e.AuthCode,
		FormationId: int32(e.FormationID),
		Map:         fieldMap(e.FieldID, ""),
		Pos:         fieldPos(e.Pos),
		Passive:     e.Passive,
		BtType:      int32(e.BtType),
	}
}

func (enc EncGame) BattleEnd(e *gmevt.BattleEnd) *field.S2C_BattleEnd {
	if e == nil {
		return nil
	}
	return &field.S2C_BattleEnd{
		Id:         e.ID,
		Won:        e.Won,
		Loots:      enc.Loots(e.Loots),
		GuildLoots: enc.Loots(e.GuildLoots),
		BtType:     int32(e.BtType),
	}
}

func (enc EncGame) Loots(ls []gmevt.Loot) []*shared.Loot {
	rv := make([]*shared.Loot, len(ls))
	for i, l := range ls {
		rv[i] = enc.Loot(l)
	}
	return rv
}

func (enc EncGame) Loot(l gmevt.Loot) *shared.Loot {
	return &shared.Loot{
		Type: shared.Loot_LootType(l.Type),
		Id:   int32(l.ID),
		Num:  int32(l.Num),
	}
}

// 角色基础数据
// func (EncGame) PlayerVer(p *gmevt.Player) *shared.Version {
// 	if p == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(p.Rev)
// 	return encGame.Version(&v)
// }

func charmToProto(c gmevt.Charm) *shared.Charm {
	return &shared.Charm{
		Num:         int32(c.Num),
		UseDailyNum: int32(c.UseDailyNum),
		LastUseTime: int32(c.LastUseTime),
	}
}

func sportRewardToProto(sr gmevt.SportReward) *shared.SportReward {
	return &shared.SportReward{
		Point:         int32(sr.Point),
		GainDailyNum:  int32(sr.GainDailyNum),
		GainWeeklyNum: int32(sr.GainWeeklyNum),
	}
}

func (enc EncGame) Player(p *gmevt.Player) *shared.Player {
	if p == nil {
		return nil
	}
	return &shared.Player{
		Uid:   int64(p.UID),
		Name:  p.Name,
		UidIM: p.UIDIM,

		Level: int32(p.Level),
		Exp:   int32(p.Exp),

		RaceID:      int32(p.RaceID),
		ClassID:     int32(p.ClassID),
		Gender:      int32(p.Gender),
		FaceID:      int32(p.FaceID),
		HairID:      int32(p.HairID),
		HairColorID: int32(p.HairColorID),

		Gold: int32(p.Gold),
		Coin: int32(p.Coin),

		Luck:             int32(p.Luck),
		CombatPower:      int32(p.CombatPower),
		Power:            int32(p.Power),
		PowerRestoreTime: p.PowerRestoreTime,
		Honour:           int32(p.Honour),
		HonourPeriodNum:  int32(p.HonourPeriodNum),
		Medal:            int32(p.Medal),
		MedalPeriodNum:   int32(p.MedalPeriodNum),

		SkillPoint:  int32(p.SkillPoint),
		TalentPoint: int32(p.TalentPoint),

		MountUID:     p.MountUID,
		Achievement:  int32(p.AchPoint),
		RunePage:     int32(p.RunePage),
		RuneResetNum: int32(p.RuneResetNum),
		BadgeID:      int32(p.BadgeID),
		// JewelChip:               int32(p.JewelChip),
		PartnerUID:              p.PartnerUID,
		ReceiveDailyBanquetTime: p.ReceiveDailyBanquetTime,
		SkillChooseID:           int32(p.SkillChooseID),
		Charm:                   charmToProto(p.Charm),
		SportReward:             sportRewardToProto(p.SportReward),

		// PrevFieldID: int32(p.PrevFieldID),
		FieldID:  int32(p.FieldID),
		Position: int32(p.Position),
	}
}

// 任务
// func (EncGame) PlayerQuestsVer(q *gmevt.PlayerQuests) *shared.Version {
// 	if q == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(q.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Quests(pq *gmevt.PlayerQuests) []*shared.Quest {
	if pq == nil {
		return nil
	}
	rv := []*shared.Quest{}
	for _, q := range pq.Quests {
		rv = append(rv, encGame.Quest(q))
	}
	return rv
}

func (EncGame) Quest(q *gmevt.Quest) *shared.Quest {
	if q == nil {
		return nil
	}
	var progress *shared.Progress
	if q.Total != 0 {
		progress = &shared.Progress{Total: int32(q.Total), Done: int32(q.Progress)}
	}
	return &shared.Quest{
		Uid:          int64(q.ID),
		Type:         shared.QuestType(q.Type),
		Status:       shared.Quest_Status(q.Status),
		Progress:     progress,
		AccomplishAt: q.AccomplishAt.Unix(),
	}
}

func (EncGame) NPCs(pn *gmevt.PlayerNPCs) []*shared.NPC {
	if pn == nil {
		return nil
	}
	rv := []*shared.NPC{}
	for _, n := range pn.NPCs {
		rv = append(rv, encGame.NPC(n))
	}
	return rv
}

func (EncGame) NPC(n *gmevt.NPC) *shared.NPC {
	if n == nil {
		return nil
	}
	return &shared.NPC{
		Uid: n.UID,
		Tid: int32(n.ID),
	}
}

func (EncGame) Item(m *gmevt.Item) *shared.Item {
	if m == nil {
		return nil
	}
	return &shared.Item{
		Uid:    m.UID,
		Tid:    int32(m.ID),
		Amount: int32(m.Amount),
		Param1: int32(m.Param1),
		Clock:  int32(m.T.Unix()),
	}
}

func (EncGame) PaperComposite(m *gmevt.PaperComposite) *shared.PaperComposite {
	if m == nil {
		return nil
	}
	return &shared.PaperComposite{
		Uid: m.UID,
		Tid: int32(m.ID),
	}
}

//*****************equipconfig begin
/*
func EquipAffix(affixes []*gmevt.EquipAffix) []*shared.EquipAffix {
	rv := []*shared.EquipAffix{}
	for _, v := range affixes {
		rv = append(rv, &shared.EquipAffix{
			Affix:    int32(v.AffixID),
			Level:    int32(v.Level),
			StatePar: int32(v.StatePar),
			Srr:      int32(v.Srr),
		})
	}
	return rv
}
*/
//*****************equipconfig end

// 面具
// func (EncGame) PlayerMasksVer(w *gmevt.PlayerMasks) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

// func (EncGame) Masks(pm *gmevt.PlayerMasks) []*shared.Mask {
// 	if pm == nil {
// 		return nil
// 	}
// 	rv := []*shared.Mask{}
// 	for _, n := range pm.Masks {
// 		rv = append(rv, encGame.Mask(n))
// 	}
// 	return rv
// }

// func (EncGame) Mask(m *gmevt.Mask) *shared.Mask {
// 	if m == nil {
// 		return nil
// 	}

// 	return &shared.Mask{
// 		Uid:        m.UID),
// 		Tid:        int32(m.ID),
// 		PartStates: m.PartStates,
// 	}
// }

// 好友
// func (EncGame) PlayerFriendsVer(w *gmevt.PlayerFriends) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Friends(pm *gmevt.PlayerFriends) []*shared.Friend {
	if pm == nil {
		return nil
	}
	rv := []*shared.Friend{}
	for _, n := range pm.Friends {
		rv = append(rv, encGame.Friend(n))
	}
	return rv
}

func UserBaseInfoToProto(u gmevt.UserBaseInfo) *base.UserBaseInfo {
	fashionIDs := []int32{}
	for _, id := range u.FashionIDs {
		fashionIDs = append(fashionIDs, int32(id))
	}
	return &base.UserBaseInfo{
		Uid:          int64(u.UID),
		Name:         u.Name,
		Lv:           int32(u.Lv),
		RaceID:       int32(u.RaceID),
		ClassID:      int32(u.ClassID),
		SpecialistID: int32(u.SpecialistID),
		Gender:       int32(u.Gender),
		FaceID:       int32(u.FaceID),
		HairID:       int32(u.HairID),
		HairColorID:  int32(u.HairColorID),
		UidIM:        u.UIDIM,
		FashionIDs:   fashionIDs,
		MountID:      int32(u.MountID),
		BadgeID:      int32(u.BadgeID),
		GuildName:    u.GuildName,
		GuildID:      u.GuildID,
		ServerID:     int32(u.ServerID),
		CombatPower:  int32(u.CombatPower),
	}
}
func UserBaseInfoPtrToProto(u *user.BaseInfo) *base.UserBaseInfo {
	return UserBaseInfoToProto(gmevt.UserBaseInfo(*u))
}
func (EncGame) Friend(m *gmevt.Friend) *shared.Friend {
	if m == nil {
		return nil
	}
	return &shared.Friend{
		UserBaseInfo: UserBaseInfoToProto(m.UserBaseInfo),
		Online:       m.Online,
	}
}

// 黑名单
// func (EncGame) PlayerBlacklistVer(w *gmevt.PlayerBlacklist) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Blacklists(pm *gmevt.PlayerBlacklist) []*shared.Blacklist {
	if pm == nil {
		return nil
	}
	rv := []*shared.Blacklist{}
	for _, n := range pm.Blacklist {
		rv = append(rv, encGame.Blacklist(n))
	}
	return rv
}

func (EncGame) Blacklist(m *gmevt.Blacklist) *shared.Blacklist {
	if m == nil {
		return nil
	}
	return &shared.Blacklist{
		UserBaseInfo: UserBaseInfoToProto(m.UserBaseInfo),
		// Uid:      proto.Int64(int64(m.UID)),
		// Name:     m.Name),
		// Level:    int32(m.Level),
		// MaskId:   int32(m.MaskID),
		// WeaponId: int32(m.WeaponID),
		// HeadId:   int32(m.HeadID),
		Online: m.Online,
		// UidIM:    m.UidIM),
		// ClassId:  int32(m.ClassId),
		// RaceId:   int32(m.RaceId),
	}
}

// 坐骑
// func (EncGame) PlayerMountsVer(w *gmevt.PlayerMounts) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Mounts(pm *gmevt.PlayerMounts) []*shared.Mount {
	if pm == nil {
		return nil
	}
	rv := []*shared.Mount{}
	for _, n := range pm.Mounts {
		rv = append(rv, encGame.Mount(n))
	}
	return rv
}

func (EncGame) Mount(m *gmevt.Mount) *shared.Mount {
	if m == nil {
		return nil
	}
	return &shared.Mount{
		Uid:     m.UID,
		Tid:     int32(m.ID),
		LevelID: int32(m.RankID),
		StarID:  int32(m.StarID),
	}
}

func (EncGame) ActivatedMounts(pm *gmevt.PlayerActivatedMounts) []*shared.ActivatedMount {
	if pm == nil {
		return nil
	}
	rv := []*shared.ActivatedMount{}
	for _, n := range pm.ActivatedMounts {
		rv = append(rv, encGame.ActivatedMount(n))
	}
	return rv
}

func (EncGame) ActivatedMount(m *gmevt.ActivatedMount) *shared.ActivatedMount {
	if m == nil {
		return nil
	}
	return &shared.ActivatedMount{
		Mount: &shared.Mount{
			Uid:     m.Mount.UID,
			Tid:     int32(m.Mount.ID),
			LevelID: int32(m.Mount.RankID),
			StarID:  int32(m.Mount.StarID),
		},
		Index: int32(m.Index),
	}
}

// 佣兵
// func (EncGame) PlayerPartnersVer(w *gmevt.PlayerPartners) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Partners(pm *gmevt.PlayerPartners) []*shared.Partner {
	if pm == nil {
		return nil
	}
	rv := []*shared.Partner{}
	for _, n := range pm.Partners {
		rv = append(rv, encGame.Partner(n))
	}
	return rv
}

func (EncGame) Partner(m *gmevt.Partner) *shared.Partner {
	if m == nil {
		return nil
	}
	return &shared.Partner{
		Uid: m.UID,
		Tid: int32(m.ID),
	}
}

// 天赋
// func (EncGame) PlayerTalentsVer(w *gmevt.PlayerTalents) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Talents(pm *gmevt.PlayerTalents) []*shared.Talent {
	if pm == nil {
		return nil
	}
	rv := []*shared.Talent{}
	for _, n := range pm.Talents {
		rv = append(rv, encGame.Talent(n))
	}
	return rv
}

func (EncGame) Talent(m *gmevt.Talent) *shared.Talent {
	if m == nil {
		return nil
	}
	return &shared.Talent{
		Uid: int32(m.UID),
	}
}

// 技能
// func (EncGame) PlayerSkillsVer(w *gmevt.PlayerSkills) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Skills(pm *gmevt.PlayerSkills) []*shared.Skill {
	if pm == nil {
		return nil
	}
	rv := []*shared.Skill{}
	for _, n := range pm.Skills {
		rv = append(rv, encGame.Skill(n))
	}
	return rv
}

func (EncGame) Skill(m *gmevt.Skill) *shared.Skill {
	if m == nil {
		return nil
	}
	return &shared.Skill{
		Uid: int32(m.UID),
	}
}

// 技能方案
// func (EncGame) PlayerSkillChoosesVer(w *gmevt.PlayerSkillChooses) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) SkillChooses(pm *gmevt.PlayerSkillChooses) []*shared.SkillChoose {
	if pm == nil {
		return nil
	}
	rv := []*shared.SkillChoose{}
	for _, n := range pm.SkillChooses {
		rv = append(rv, encGame.SkillChoose(n))
	}
	return rv
}

func (enc EncGame) SkillChoose(m *gmevt.SkillChoose) *shared.SkillChoose {
	if m == nil {
		return nil
	}
	return &shared.SkillChoose{
		Id:      int32(m.ID),
		TeamIds: intutil.IntToInt32Slice(m.TeamIDs),
	}
}

// 装备
// func (EncGame) PlayerEquipsVer(w *gmevt.PlayerEquips) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Equipments(pm *gmevt.PlayerEquips) []*shared.Equipment {
	if pm == nil {
		return nil
	}

	rv := []*shared.Equipment{}
	for _, n := range pm.Equips {
		rv = append(rv, encGame.Equipment(n))
	}

	return rv
}

//*****************equipconfig begin
func (enc EncGame) Equipment(m *gmevt.Equipment) *shared.Equipment {
	if m == nil {
		return nil
	}
	return &shared.Equipment{
		Uid: m.UID,
		Tid: int32(m.ID),
		//Affixes:         EquipAffix(m.Affixes),
		//Skills:          protoIntSlice(m.Skills),
		//RecastTimes:     int32(m.RecastTimes),
		Bind: int32(m.Bind),
		//LockedJewelSlot: enc.JewelSlot(m.LockedJewelSlot),
		JewelSlots: enc.JewelSlots(m.JewelSlots),
		ExtraProps: enc.ExtraProps(m.ExtraProps),
		//ShiftSkills:     enc.ShiftSkills(m.ShiftSkills),
		FashionID: int32(m.FashionID),
	}
}

//*****************equipconfig end

func (enc EncGame) ShiftSkill(skill *gmevt.ShiftSkill) *shared.ShiftSkill {
	if skill == nil {
		return nil
	}

	return &shared.ShiftSkill{
		Id: int32(skill.ID),
	}
}

func (enc EncGame) ShiftSkills(skills []*gmevt.ShiftSkill) []*shared.ShiftSkill {
	rv := []*shared.ShiftSkill{}
	for _, skill := range skills {
		if skill != nil {
			rv = append(rv, enc.ShiftSkill(skill))
		}
	}
	return rv
}

func (enc EncGame) ExtraProp(ep *gmevt.ExtraProp) *shared.ExtraProp {
	if ep == nil {
		return nil
	}

	return &shared.ExtraProp{
		Index:   int32(ep.Index),
		PropID:  int32(ep.PropID),
		Quality: int32(ep.Quality),
		Val:     int32(ep.Val),
	}
}

func (enc EncGame) ExtraProps(extraProps []*gmevt.ExtraProp) []*shared.ExtraProp {
	rv := []*shared.ExtraProp{}
	for _, ep := range extraProps {
		if ep != nil {
			rv = append(rv, enc.ExtraProp(ep))
		}
	}

	return rv
}

func (enc EncGame) JewelSlot(slot *gmevt.JewelSlot) *shared.JewelSlot {
	if slot == nil {
		return nil
	}
	return &shared.JewelSlot{
		Jewel: enc.Jewel(slot.Jewel),
		//UnlockCond: int32(slot.UnlockCond),
		//Team:       int32(slot.Team),
		JewelType: slot.JewelSlotType,
	}
}

func (enc EncGame) JewelSlots(slots []*gmevt.JewelSlot) []*shared.JewelSlot {
	rv := []*shared.JewelSlot{}
	for _, slot := range slots {
		if slot != nil {
			rv = append(rv, enc.JewelSlot(slot))
		}
	}
	return rv
}

// 宝石
// func (EncGame) PlayerJewelsVer(w *gmevt.PlayerJewels) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

// func (EncGame) PlayerActiveRewardsVer(w *gmevt.PlayerActiveRewards) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

// func (EncGame) PlayerFinishFormationInfoVer(w *gmevt.PlayerFinishFormationInfo) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (enc EncGame) Jewels(pm *gmevt.PlayerJewels) []*shared.Jewel {
	if pm == nil {
		return nil
	}
	rv := []*shared.Jewel{}
	for _, n := range pm.Jewels {
		rv = append(rv, enc.Jewel(n))
	}
	return rv
}

//*****************equipconfig begin
func (EncGame) Jewel(m *gmevt.Jewel) *shared.Jewel {
	if m == nil {
		return nil
	}
	return &shared.Jewel{
		Uid: m.UID,
		Tid: int32(m.ID),
		//		Level:       int32(m.Level),
		Exp: int32(m.Exp),
		//Affixes:     EquipAffix(m.Affixes),
		//RecastTimes: int32(m.RecastTimes),
		Bind: int32(m.Bind),
	}
}

//*****************equipconfig end

func (enc EncGame) ActiveReward(ar *gmevt.ActiveReward) *shared.ActiveReward {
	if ar == nil {
		return nil
	}
	gainInfo := []*shared.GainInfo{}
	for level, t := range ar.GainInfo {
		gainInfo = append(gainInfo, &shared.GainInfo{
			Level: int32(level),
			Clock: int32(t.Unix()),
		})
	}
	return &shared.ActiveReward{
		Type:       int32(ar.Type),
		Point:      int32(ar.Point),
		PointClock: int32(ar.PointTime.Unix()),
		GainInfos:  gainInfo,
	}
}

func (enc EncGame) ActiveRewards(ar *gmevt.PlayerActiveRewards) []*shared.ActiveReward {
	if ar == nil {
		return nil
	}
	ret := []*shared.ActiveReward{}
	for _, v := range ar.ActiveRewards {
		ret = append(ret, enc.ActiveReward(v))
	}
	return ret
}

func (enc EncGame) FinishFormationInfo(ffInfo *gmevt.PlayerFinishFormationInfo) *shared.FinishFormationInfo {
	if ffInfo == nil {
		return nil
	}
	raidFFInfos := []*shared.RaidFFInfo{}
	for _, mapFF := range ffInfo.FinishFormationInfo.RaidFFs {
		finishFormations := []*shared.FinishFormation{}
		for _, ff := range mapFF.FinishFormations {
			finishFormations = append(finishFormations, &shared.FinishFormation{
				Id:    int32(ff.ID),
				Num:   int32(ff.Num),
				Clock: int32(ff.Time.Unix()),
			})
		}
		raidFFInfos = append(raidFFInfos, &shared.RaidFFInfo{
			MapID:            int32(mapFF.MapID),
			FinishFormations: finishFormations,
		})
	}

	return &shared.FinishFormationInfo{RaidFFInfos: raidFFInfos}
}

func (enc EncGame) FinishFormation(ff *gmevt.FinishFormation) *shared.FinishFormation {
	if ff == nil {
		return nil
	}
	return &shared.FinishFormation{
		Id:    int32(ff.ID),
		Num:   int32(ff.Num),
		Clock: int32(ff.Time.Unix()),
	}
}

func (enc EncGame) SubMapFID(subMapFID *gmevt.SubMapFID) *shared.SubMapFID {
	if subMapFID == nil {
		return nil
	}
	fIDs := []int32{}
	for _, fID := range subMapFID.FIDs {
		fIDs = append(fIDs, int32(fID))
	}
	return &shared.SubMapFID{
		SubMapID: int32(subMapFID.SubMapID),
		FIDs:     fIDs,
	}
}

func (enc EncGame) SubMapFFID(subMapFFID *gmevt.SubMapFFID) *shared.SubMapFFID {
	if subMapFFID == nil {
		return nil
	}
	finishFormationIDs := []int32{}
	for _, ffID := range subMapFFID.FinishFormationIDs {
		finishFormationIDs = append(finishFormationIDs, int32(ffID))
	}
	return &shared.SubMapFFID{
		SubMapID:           int32(subMapFFID.SubMapID),
		FinishFormationIDs: finishFormationIDs,
	}
}
func (enc EncGame) SubMapBFID(subMapBFID *gmevt.SubMapBFID) *shared.SubMapBFID {
	if subMapBFID == nil {
		return nil
	}
	battleFormationIDs := []int32{}
	for _, bfID := range subMapBFID.BattleFormationIDs {
		battleFormationIDs = append(battleFormationIDs, int32(bfID))
	}
	return &shared.SubMapBFID{
		SubMapID:           int32(subMapBFID.SubMapID),
		BattleFormationIDs: battleFormationIDs,
	}
}

func (enc EncGame) SubMapFFIDs(subMapFFIDs []*gmevt.SubMapFFID) []*shared.SubMapFFID {
	if len(subMapFFIDs) == 0 {
		return []*shared.SubMapFFID{}
	}
	ret := []*shared.SubMapFFID{}
	for _, subMapFFID := range subMapFFIDs {
		ret = append(ret, enc.SubMapFFID(subMapFFID))
	}
	return ret
}
func (enc EncGame) SubMapFIDs(subMapFIDs []*gmevt.SubMapFID) []*shared.SubMapFID {
	if len(subMapFIDs) == 0 {
		return []*shared.SubMapFID{}
	}
	ret := []*shared.SubMapFID{}
	for _, subMapFID := range subMapFIDs {
		ret = append(ret, enc.SubMapFID(subMapFID))
	}
	return ret
}

func (enc EncGame) SubMapBFIDs(subMapBFIDs []*gmevt.SubMapBFID) []*shared.SubMapBFID {
	if len(subMapBFIDs) == 0 {
		return []*shared.SubMapBFID{}
	}
	ret := []*shared.SubMapBFID{}
	for _, subMapBFID := range subMapBFIDs {
		ret = append(ret, enc.SubMapBFID(subMapBFID))
	}
	return ret
}

func (enc EncGame) Raid(r *gmevt.Raid) *shared.Raid {
	if r == nil {
		return nil
	}
	finishFormations := []*shared.FinishFormation{}
	for _, ff := range r.FinishFormations {
		finishFormations = append(finishFormations, &shared.FinishFormation{
			Id:    int32(ff.ID),
			Num:   int32(ff.Num),
			Clock: int32(ff.Time.Unix()),
		})
	}

	return &shared.Raid{
		MapID:            int32(r.MapID),
		FinishFormations: finishFormations,
		SubMapFFIDs:      enc.SubMapFFIDs(r.SubMapFFIDs),
		SubMapBFIDs:      enc.SubMapBFIDs(r.SubMapBFIDs),
		CreateClock:      int32(r.CreateTime.Unix()),
	}
}

func (enc EncGame) Dungeon(d *gmevt.Dungeon) *shared.Dungeon {
	if d == nil {
		return nil
	}
	return &shared.Dungeon{
		MapID:       int32(d.MapID),
		SubMapFFIDs: enc.SubMapFFIDs(d.SubMapFFIDs),
		SubMapBFIDs: enc.SubMapBFIDs(d.SubMapBFIDs),
		SubMapRFIDs: enc.SubMapFIDs(d.SubMapRFIDs),
		CreateClock: int32(d.CreateClock),
		EndClock:    int32(d.EndClock),
		BFinished:   d.BFinished,
	}
}

func (enc EncGame) Wonderland(w *gmevt.Wonderland) *shared.Wonderland {
	if w == nil {
		return nil
	}
	specBuffIDs := []int32{}
	for _, id := range w.SpecBuffIDs {
		specBuffIDs = append(specBuffIDs, int32(id))
	}
	return &shared.Wonderland{
		ID:          int32(w.ID),
		KeyUID:      int32(w.KeyUID),
		KeyType:     int32(w.KeyType),
		Layer:       int32(w.Layer),
		SubMapFFIDs: enc.SubMapFFIDs(w.SubMapFFIDs),
		SubMapBFIDs: enc.SubMapBFIDs(w.SubMapBFIDs),
		StartTime:   int32(w.StartTime),
		Duration:    int32(w.Duration),
		CloseReason: int32(w.CloseReason),
		Star:        int32(w.Star),
		BuffID:      int32(w.BuffID),
		SpecBuffIDs: specBuffIDs,
	}
}

// 工具
// func (EncGame) PlayerToolsVer(w *gmevt.PlayerTools) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Tools(pm *gmevt.PlayerTools) []*shared.Tool {
	if pm == nil {
		return nil
	}
	rv := []*shared.Tool{}
	for _, n := range pm.Tools {
		rv = append(rv, encGame.Tool(n))
	}
	return rv
}

func (EncGame) Tool(m *gmevt.Tool) *shared.Tool {
	if m == nil {
		return nil
	}
	return &shared.Tool{
		Tid: int32(m.ID),
	}
}

// func (EncGame) PlayerCouponsVer(w *gmevt.PlayerCoupons) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

// func (EncGame) PlayerPaymentsVer(w *gmevt.PlayerPayments) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Coupons(pm *gmevt.PlayerCoupons) []*shared.Coupon {
	if pm == nil {
		return nil
	}
	rv := []*shared.Coupon{}
	for _, n := range pm.Coupons {
		rv = append(rv, encGame.Coupon(n))
	}
	return rv
}

func (EncGame) Coupon(m *gmevt.Coupon) *shared.Coupon {
	if m == nil {
		return nil
	}
	return &shared.Coupon{
		Uid:        m.UniqueID,
		CouponId:   int32(m.CouponID),
		FromUserId: int32(m.FromUID),
		CreatedAt:  m.CreatedAt.Unix(),
		ExpireAt:   m.ExpireAt.Unix(),
	}
}

func (EncGame) Payments(pm *gmevt.PlayerPayments) []*shared.Payment {
	if pm == nil {
		return nil
	}
	rv := []*shared.Payment{}
	for _, n := range pm.Payments {
		rv = append(rv, encGame.Payment(n))
	}
	return rv
}

func (EncGame) Payment(m *gmevt.Payment) *shared.Payment {
	if m == nil {
		return nil
	}
	return &shared.Payment{
		PayKey:    shared.Payment_PayKey(m.Key),
		Value:     int32(m.Value),
		CreatedAt: m.CreatedAt.Unix(),
		ExpireAt:  m.ExpireAt.Unix(),
	}
}

// kv store
func (EncGame) KVStores(m *gmevt.PlayerKVStores) *shared.KVStores {
	if m == nil {
		return nil
	}
	rv := &shared.KVStores{}

	if generic, ok := (*m.KVStores)[strconv.Itoa(int(kvstore.KVCategory_generic))]; ok {
		rv.Generic = &kvstore.KVStore{
			Items: kvItems(generic),
		}
	}

	if tutorial, ok := (*m.KVStores)[strconv.Itoa(int(kvstore.KVCategory_tutorial))]; ok {
		rv.Tutorial = &kvstore.KVStore{
			Items: kvItems(tutorial),
		}
	}

	return rv
}

// 背包扩展
// func (EncGame) PlayerBagExtendsVer(w *gmevt.PlayerBagExtends) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) BagExtends(pm *gmevt.PlayerBagExtends) []*shared.BagExtend {
	if pm == nil {
		return nil
	}
	rv := []*shared.BagExtend{}
	for _, n := range pm.BagExtends {
		rv = append(rv, encGame.BagExtend(n))
	}
	return rv
}

func (EncGame) BagExtend(m *gmevt.BagExtend) *shared.BagExtend {
	if m == nil {
		return nil
	}
	return &shared.BagExtend{
		Type:  int32(m.Type),
		Level: int32(m.Level),
	}
}

func kvItems(m *gmevt.KVStore) []*kvstore.KVItem {
	rv := make([]*kvstore.KVItem, 0, len(*m))
	for k, v := range *m {
		rv = append(rv, &kvstore.KVItem{
			Key:   k,
			Value: v,
		})
	}
	return rv
}

// func (EncGame) PlayerKVStoresVer(m *gmevt.PlayerKVStores) *shared.Version {
// 	if m == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(m.Rev)
// 	return encGame.Version(&v)
// }

// 声望
// func (EncGame) PlayerFamesVer(w *gmevt.PlayerFames) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Fames(pm *gmevt.PlayerFames) []*shared.Fame {
	if pm == nil {
		return nil
	}
	rv := []*shared.Fame{}
	for _, n := range pm.Fames {
		rv = append(rv, encGame.Fame(n))
	}
	return rv
}

func (EncGame) Fame(m *gmevt.Fame) *shared.Fame {
	if m == nil {
		return nil
	}
	return &shared.Fame{
		Uid:   int32(m.UID),
		Level: int32(m.Level),
		Exp:   int32(m.Exp),
	}
}

// 生活技能
// func (EncGame) PlayerLifeSkillsVer(w *gmevt.PlayerLifeSkills) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) LifeSkills(pm *gmevt.PlayerLifeSkills) []*shared.LifeSkill {
	if pm == nil {
		return nil
	}
	rv := []*shared.LifeSkill{}
	for _, n := range pm.LifeSkills {
		rv = append(rv, encGame.LifeSkill(n))
	}
	return rv
}

func (EncGame) LifeSkill(m *gmevt.LifeSkill) *shared.LifeSkill {
	if m == nil {
		return nil
	}
	return &shared.LifeSkill{
		Uid:   int32(m.UID),
		Level: int32(m.Level),
		Exp:   int32(m.Exp),
	}
}

// 地图buff
// func (EncGame) PlayerMapBuffVer(w *gmevt.PlayerMapBuffs) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) MapBuffs(pm *gmevt.PlayerMapBuffs) []*shared.MapBuff {
	if pm == nil {
		return nil
	}
	rv := []*shared.MapBuff{}
	for _, n := range pm.MapBuffs {
		rv = append(rv, encGame.MapBuff(n))
	}
	return rv
}

func (EncGame) MapBuff(m *gmevt.MapBuff) *shared.MapBuff {
	if m == nil {
		return nil
	}
	return &shared.MapBuff{
		Uid:         m.UID,
		Tid:         int32(m.ID),
		RemainTime:  int32(m.RemainTime),
		DestroyTime: m.DestroyTime.Unix(),
	}
}

// 符文
// func (EncGame) PlayerRuneVer(w *gmevt.PlayerRunes) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Runes(pm *gmevt.PlayerRunes) []*shared.Rune {
	if pm == nil {
		return nil
	}
	rv := []*shared.Rune{}
	for _, n := range pm.Runes {
		rv = append(rv, encGame.Rune(n))
	}
	return rv
}

func (EncGame) Rune(m *gmevt.Rune) *shared.Rune {
	if m == nil {
		return nil
	}
	return &shared.Rune{
		Tid:    int32(m.ID),
		Amount: int32(m.Amount),
	}
}

func (EncGame) Money(m *gmevt.Money) *shared.Money {
	if m == nil {
		return nil
	}
	return &shared.Money{
		Tid:    int32(m.ID),
		Amount: int32(m.Amount),
	}
}

// func (EncGame) PlayerSignVer(w *gmevt.PlayerSigns) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

// 符文方案
// func (EncGame) PlayerRunePageVer(w *gmevt.PlayerRunePages) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) RunePages(pm *gmevt.PlayerRunePages) []*shared.RunePage {
	if pm == nil {
		return nil
	}
	rv := []*shared.RunePage{}
	for _, n := range pm.RunePages {
		rv = append(rv, encGame.RunePage(n))
	}
	return rv
}

func (EncGame) RunePage(m *gmevt.RunePage) *shared.RunePage {
	if m == nil {
		return nil
	}
	return &shared.RunePage{
		Tid:   int32(m.ID),
		Runes: protoIntSlice(m.Runes),
	}
}

// 签到
func (EncGame) Signs(e *gmevt.PlayerSigns) []*shared.Sign {
	if e == nil {
		return nil
	}
	rv := []*shared.Sign{}
	for _, data := range e.Signs {
		rv = append(rv, encGame.Sign(&data))
	}
	return rv
}

func (EncGame) Sign(e *gmevt.Sign) *shared.Sign {
	if e == nil {
		return nil
	}
	return &shared.Sign{
		Year:        int32(e.Year),
		Month:       int32(e.Month),
		Day:         int32(e.Day),
		BSign:       e.BSign,
		BAlreadyGot: e.BAlreadyGot,
	}
}

// 背包
// func (EncGame) PlayerBagVer(w *gmevt.PlayerBag) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) Bag(pm *gmevt.PlayerBag) *shared.Bag {
	if pm == nil {
		return nil
	}
	rv := shared.Bag{}
	rv.Cells = make([]*shared.ItemCell, 0)
	for _, n := range pm.Cells {
		rv.Cells = append(rv.Cells, encGame.ItemCell(n))
	}
	return &rv
}

func (enc EncGame) ItemCell(m *gmevt.ItemCell) *shared.ItemCell {
	if m == nil {
		return nil
	}
	rv := shared.ItemCell{}

	if m.GetItem() != nil {
		cell := &shared.ItemCell_Item{}
		cell.Item = enc.Item(m.GetItem())
		rv.Content = cell
	}

	if m.GetEquip() != nil {
		cell := &shared.ItemCell_Equip{}
		cell.Equip = enc.Equipment(m.GetEquip())
		rv.Content = cell
	}

	if m.GetPaperComposite() != nil {
		cell := &shared.ItemCell_PaperComposite{}
		cell.PaperComposite = enc.PaperComposite(m.GetPaperComposite())
		rv.Content = cell
	}

	return &rv
}

// func (EncGame) PlayerFormulaIDVer(w *gmevt.PlayerFormulaIDs) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

// 成就
// func (EncGame) PlayerAchievementVer(w *gmevt.PlayerAchievements) *shared.Version {
// 	if w == nil {
// 		return nil
// 	}
// 	v := gmevt.Version(w.Rev)
// 	return encGame.Version(&v)
// }

func (EncGame) PlayerAchievements(pm *gmevt.PlayerAchievements) []*shared.Achievement {
	if pm == nil {
		return nil
	}
	rv := []*shared.Achievement{}
	for _, n := range pm.Achievements {
		rv = append(rv, encGame.Achievement(n))
	}
	return rv
}

func (EncGame) Achievements(as []*gmevt.Achievement) []*shared.Achievement {
	rv := []*shared.Achievement{}
	for _, n := range as {
		rv = append(rv, encGame.Achievement(n))
	}
	return rv
}

func (EncGame) Achievement(m *gmevt.Achievement) *shared.Achievement {
	if m == nil {
		return nil
	}
	return &shared.Achievement{
		Uid:          int32(m.UID),
		Progress:     int32(m.Progress),
		State:        int32(m.State),
		FinishedTime: m.FinishedTime.Unix(),
	}
}

func (EncGame) FormulaIDs(pm *gmevt.PlayerFormulaIDs) []*shared.FormulaId {
	if pm == nil {
		return nil
	}
	rv := []*shared.FormulaId{}
	for _, n := range pm.FormulaIDs {
		rv = append(rv, encGame.FormulaID(&n))
	}
	return rv
}

func (EncGame) FormulaID(m *gmevt.FormulaID) *shared.FormulaId {
	if m == nil {
		return nil
	}
	return &shared.FormulaId{
		Id: int32(m.ID),
	}
}

// version
// func (EncGame) Version(v *gmevt.Version) *shared.Version {
// 	if v == nil {
// 		return nil
// 	}
// 	ver := int(*v)
// 	return &shared.Version{V: int32(ver)}
// }

func (EncGame) CsnChatMessage(m *gmevt.CsnChatMessage) *chat.CSN_ChatMessage {
	if m == nil {
		return nil
	}
	return &chat.CSN_ChatMessage{
		ChatId:      m.ChatId,
		ChatType:    m.ChatType,
		Msg:         m.Msg,
		FromId:      m.FromId,
		FromName:    m.FromName,
		AudioStream: m.AudioStream,
	}
}

func (enc EncGame) ResUpdateFormation(m *gmevt.ResUpdateFormation) *action.S2C_UpdateFormation {
	if m == nil {
		return nil
	}
	return &action.S2C_UpdateFormation{}
}

func getChatId(chatID gmevt.ChatID) chat.EChatID {
	var val chat.EChatID
	switch chatID {
	case gmevt.Chat_World:
		val = chat.EChatID_world
	case gmevt.Chat_Guild:
		val = chat.EChatID_guild
	case gmevt.Chat_Team:
		val = chat.EChatID_team
	case gmevt.Chat_Local:
		val = chat.EChatID_local
	case gmevt.Chat_Personal:
		val = chat.EChatID_personal
	case gmevt.Chat_TeamReq:
		val = chat.EChatID_teamreq
	}
	return val
}

func (EncGame) ResUpdateChatInfo(m *gmevt.ResUpdateChatInfo) *chat.S2C_UpdateChatInfo {
	if m == nil {
		return nil
	}
	infos := []*chat.ChatInfo{}
	for _, v := range m.ChatInfos {
		infos = append(infos, &chat.ChatInfo{
			ChatId:     getChatId(v.ChatID),
			ChannelUID: int32(v.UID),
		})
	}
	return &chat.S2C_UpdateChatInfo{ChatInfos: infos}
}

func (enc EncGame) ResEquipMask(m *gmevt.ResEquipMask) *action.S2C_EquipMask {
	if m == nil {
		return nil
	}
	return &action.S2C_EquipMask{}
}

func (enc EncGame) ResEquipEquipment(m *gmevt.ResEquipEquipment) *action.S2C_EquipEquipment {
	if m == nil {
		return nil
	}
	return &action.S2C_EquipEquipment{
		Oper: enc.Operation(m.Oper),
	}
}

func (enc EncGame) ResUnEquipEquipment(m *gmevt.ResUnEquipEquipment) *action.S2C_UnEquipEquipment {
	if m == nil {
		return nil
	}
	return &action.S2C_UnEquipEquipment{
		Oper: enc.Operation(m.Oper),
	}
}

/*
func (enc EncGame) ResEquipMount(m *gmevt.ResEquipMount) *action.S2C_EquipMount {
	if m == nil {
		return nil
	}
	return &action.S2C_EquipMount{}
}

func (enc EncGame) ResUnEquipMount(m *gmevt.ResUnEquipMount) *action.S2C_UnEquipMount {
	if m == nil {
		return nil
	}
	return &action.S2C_UnEquipMount{}
}
*/

func (enc EncGame) ResEquipPartner(m *gmevt.ResEquipPartner) *action.S2C_EquipPartner {
	if m == nil {
		return nil
	}
	return &action.S2C_EquipPartner{}
}

func (enc EncGame) ResUnEquipPartner(m *gmevt.ResUnEquipPartner) *action.S2C_UnEquipPartner {
	if m == nil {
		return nil
	}
	return &action.S2C_UnEquipPartner{}
}

func (enc EncGame) ResSellItem(m *gmevt.ResSellItem) *action.S2C_SellItem {
	if m == nil {
		return nil
	}
	return &action.S2C_SellItem{}
}

func (enc EncGame) ResSortItem(m *gmevt.ResSortItem) *action.S2C_SortItem {
	if m == nil {
		return nil
	}
	return &action.S2C_SortItem{}
}

func (enc EncGame) ResUseItem(m *gmevt.ResUseItem) *action.S2C_UseItem {
	if m == nil {
		return nil
	}
	return &action.S2C_UseItem{
		Ok: m.OK,
	}
}

func friendUser(u *gmevt.FriendUser) *friend.FriendUser {
	if u == nil {
		return nil
	}
	return &friend.FriendUser{
		UserBaseInfo: UserBaseInfoToProto(u.UserBaseInfo),
		// Uid:          int64(u.UID)),
		// Name:         u.Name),
		// Level:        int32(u.Level),
		// MaskId:       int32(u.MaskID),
		// HeadId:       int32(u.HeadID),
		// WeaponId:     int32(u.WeaponID),
		// UidIM:        u.UIDIM),
		// ClassId:      int32(u.ClassId),
		// RaceId:       int32(u.RaceId),
		Online: u.Online,
		// HasCouponBox: u.HasCouponBox),
	}
}

func (enc EncGame) ResLocateUser(m *gmevt.ResLocateUser) *friend.S2C_LocateUser {
	if m == nil {
		return nil
	}
	var users []*friend.FriendUser
	for _, v := range m.Users {
		users = append(users, friendUser(v))
	}
	return &friend.S2C_LocateUser{
		Users: users,
	}
}

func (enc EncGame) ResRecommendFriend(m *gmevt.ResRecommendFriend) *friend.S2C_RecommendFriend {
	if m == nil {
		return nil
	}
	var users []*friend.FriendUser
	for _, v := range m.Users {
		users = append(users, friendUser(v))
	}
	return &friend.S2C_RecommendFriend{
		Users: users,
	}
}

func (enc EncGame) ResSearchNearbyUser(m *gmevt.ResSearchNearbyUser) *friend.S2C_SearchNearbyUser {
	if m == nil {
		return nil
	}
	var users []*friend.FriendUser
	for _, v := range m.Users {
		users = append(users, friendUser(v))
	}
	return &friend.S2C_SearchNearbyUser{
		Users: users,
	}
}

func (enc EncGame) ResAddFriendByUID(m *gmevt.ResAddFriendByUID) *friend.S2C_AddFriendByUID {
	if m == nil {
		return nil
	}
	return &friend.S2C_AddFriendByUID{
		Uid:  int64(m.UID),
		Name: m.Name,
	}
}

func (enc EncGame) ResAddFriendByName(m *gmevt.ResAddFriendByName) *friend.S2C_AddFriendByName {
	if m == nil {
		return nil
	}
	return &friend.S2C_AddFriendByName{
		Uid:  int64(m.UID),
		Name: m.Name,
	}
}

func (enc EncGame) ResRemoveFriendByUID(m *gmevt.ResRemoveFriendByUID) *friend.S2C_RemoveFriendByUID {
	if m == nil {
		return nil
	}
	return &friend.S2C_RemoveFriendByUID{
		Uid:  int64(m.UID),
		Name: m.Name,
	}
}

func (enc EncGame) ResAddBlacklist(m *gmevt.ResAddBlacklist) *friend.S2C_AddBlacklist {
	if m == nil {
		return nil
	}
	return &friend.S2C_AddBlacklist{
		Ok:   m.OK,
		Name: m.Name,
	}
}

func (enc EncGame) ResRemoveBlacklist(m *gmevt.ResRemoveBlacklist) *friend.S2C_RemoveBlacklist {
	if m == nil {
		return nil
	}
	return &friend.S2C_RemoveBlacklist{
		Ok:   m.OK,
		Name: m.Name,
	}
}

// 好友数据，只传部分字段
func friendUserWithFields(u *gmevt.FriendUser) *friend.FriendUser {
	if u == nil {
		return nil
	}
	// TODO
	return &friend.FriendUser{
		UserBaseInfo: UserBaseInfoToProto(u.UserBaseInfo),
		// Uid:          int64(u.UID)),
		// Name:         u.Name),
		// Level:        int32(u.Level),
		Online: u.Online,
		// HasCouponBox: u.HasCouponBox),
	}
}

func (enc EncGame) ResFriendStatusList(m *gmevt.ResFriendStatusList) *friend.S2C_FriendStatusList {
	if m == nil {
		return nil
	}
	var users []*friend.FriendUser
	for _, v := range m.Users {
		users = append(users, friendUserWithFields(v))
	}
	return &friend.S2C_FriendStatusList{
		Users: users,
	}
}

func (enc EncGame) ResRecentTeamMemberList(m *gmevt.ResRecentTeamMemberList) *friend.S2C_RecentTeamMemberList {
	if m == nil {
		return nil
	}
	var users []*friend.FriendUser
	for _, v := range m.Users {
		users = append(users, friendUser(v))
	}
	return &friend.S2C_RecentTeamMemberList{
		Users: users,
	}
}

// func (enc EncGame) ResAutoMatch(m *gmevt.ResAutoMatch) *team.S2C_AutoMatch {
// 	if m == nil {
// 		return nil
// 	}
// 	return &team.S2C_AutoMatch{
// 		Ok: m.OK),
// 	}
// }

////Battle messages
//
//func (EncBattle) Start(s *btevt.Start) *battle.S2C_Start {
//	params := []*battle.Param{}
//	for _, d := range s.Params {
//		params = append(params, encBattle.Param(&d))
//	}
//	return &battle.S2C_Start{
//		Id:     int32(s.Id),
//		Map:    &battle.Map{MaxX: int32(s.Ground.NX), MaxY: int32(s.Ground.NY)},
//		State:  encBattle.State(s.State),
//		Params: params,
//	}
//}
//
//func (EncBattle) Pong(e *btevt.Pong) *battle.Pong {
//	return new(battle.Pong)
//}
//
//func (EncBattle) State(s btevt.State) *battle.S2C_State {
//	players := []*battle.Player{}
//	for _, p := range s.Players {
//		players = append(players, encBattle.Player(&p))
//	}
//
//	return &battle.S2C_State{
//		Players: players,
//		AtMs:    int32(s.AtMs),
//	}
//}
//
//func (EncBattle) StartMoving(e *btevt.StartMoving) *battle.S2C_StartMoving {
//	return &battle.S2C_StartMoving{
//		FighterId:  int32(e.FighterID),
//		AtMs:       int32(e.AtMs),
//		MovePosNum: int32(e.MovePosNum),
//		BFront:     e.BFront),
//		ToPosNum:   int32(e.ToPosNum),
//		AtPosNum:   int32(e.AtPosNum),
//	}
//}
//
//func (EncBattle) EndMoving(e *btevt.EndMoving) *battle.S2C_EndMoving {
//	return &battle.S2C_EndMoving{
//		FighterId:  int32(e.FighterID),
//		AtMs:       int32(e.AtMs),
//		AtPosNum:   int32(e.AtPosNum),
//		MovingType: int32(e.MovingType),
//	}
//}
//
//func (EncBattle) Player(p *btevt.Player) *battle.Player {
//	facing := battle.Player_left
//	if p.FacingDirection == descriptor.Right {
//		facing = battle.Player_right
//	}
//	fighters := []*battle.FighterDetail{}
//	for _, u := range p.Units {
//		fighters = append(fighters, encBattle.FighterDetail(&u))
//	}
//	return &battle.Player{
//		Id:       int32(p.ID),
//		Team:     int32(p.TeamID),
//		Facing:   &facing,
//		Fighters: fighters,
//	}
//}
//
//func (EncBattle) Param(p *btevt.Param) *battle.Param {
//	return &battle.Param{
//		K: &p.K,
//		V: &p.V,
//	}
//}
//
//func (EncBattle) FighterDetail(u *btevt.Unit) *battle.FighterDetail {
//	skillSet := []int32{}
//	for _, skill := range u.SkillSet {
//		skillSet = append(skillSet, int32(skill))
//	}
//	var iid int
//	var t battle.FighterDetail_FighterType
//	if u.MaskID != 0 {
//		t = battle.FighterDetail_player
//	} else if u.BNPCID != 0 {
//		t = battle.FighterDetail_bnpc
//		iid = u.BNPCID
//	} else if u.MercID != 0 {
//		t = battle.FighterDetail_merc
//		iid = u.MercID
//	}
//
//	var bbs []*battle.Buff
//	for _, b := range u.Buffs {
//		bbs = append(bbs, encBattle.Buff(b))
//	}
//
//	skillStances := []*battle.SkillStance{}
//	for _, skillStance := range u.SkillStances {
//		skillStances = append(skillStances, &battle.SkillStance{
//			SkillId:      int32(skillStance.SkillId),
//			Stance:       int32(skillStance.Stance),
//			ClassSkillId: int32(skillStance.ClassSkillId),
//			BPrimary:     skillStance.BPrimary),
//		})
//	}
//
//	return &battle.FighterDetail{
//		Id:           int32(u.ID),
//		Type:         t.Enum(),
//		MaskId:       int32(u.MaskID),
//		EquipIds:     protoIntSlice(u.EquipIDs),
//		InstanceId:   int32(iid),
//		BossTag:      battle.FighterDetail_BossType(u.BossTag).Enum(),
//		Name:         &u.Name,
//		Hp:           int32(u.HP),
//		MaxHp:        int32(u.MaxHP),
//		AttackPower:  int32(u.AttackPower),
//		AbilityPower: int32(u.AbilityPower),
//		Armor:        int32(u.Armor),
//		MArmor:       int32(u.MArmor),
//		Location:     encBattle.Position(u.Pos),
//		SkillSet:     skillSet,
//		SquadTag:     int32(u.SquadTag),
//		Shield:       int32(u.Shield),
//		Buffs:        bbs,
//		Stance:       int32(u.Stance),
//		SkillStances: skillStances,
//		Energy:       int32(u.Energy),
//		MovingSpeed:  int32(u.MovingSpeed),
//		Lv:           int32(u.Lv),
//		HeadId:       int32(u.HeadID),
//	}
//}
//
//func (EncBattle) PleaseInput(e *btevt.PleaseInput) *battle.S2C_PleaseInputInstructions {
//	return &battle.S2C_PleaseInputInstructions{
//		Round:   int32(e.Round),
//		Timeout: int32(e.Timeout),
//	}
//}
//
//func (EncBattle) InstructionAck(e *btevt.InstructionAck) *battle.S2C_SkillInstruction {
//	return &battle.S2C_SkillInstruction{
//		SkillId:   int32(e.SkillID),
//		TargetId:  int32(e.TargetID),
//		FighterId: int32(e.FighterID),
//	}
//}

//var resultDict = map[btevt.ResultType]battle.S2C_SkillResult_Result{
//	btevt.OK:                battle.S2C_SkillResult_ok,
//	btevt.NoValidTarget:     battle.S2C_SkillResult_noValidTarget,
//	btevt.NotWithinDistance: battle.S2C_SkillResult_notWithinDistance,
//	btevt.Miss:              battle.S2C_SkillResult_miss,
//	btevt.Immunity:          battle.S2C_SkillResult_immunity,
//	btevt.CallNoFreePosNum:  battle.S2C_SkillResult_callNoFreePosNum,
//}
//
//var eventDict = map[btevt.Event]battle.BuffEvent_Event{
//	btevt.BE_Add:      battle.BuffEvent_BE_Add,
//	btevt.BE_Remove:   battle.BuffEvent_BE_Remove,
//	btevt.BE_DO:       battle.BuffEvent_BE_DO,
//	btevt.BE_Update:   battle.BuffEvent_BE_Update,
//	btevt.BE_RunAway:  battle.BuffEvent_BE_RunAway,
//	btevt.BE_Immunity: battle.BuffEvent_BE_Immunity,
//	btevt.BE_Catch:    battle.BuffEvent_BE_Catch,
//}
//
//func (EncBattle) BuffTickResult(e *btevt.BuffTickResult) *battle.S2C_TickBuff {
//	buffGroups := []*battle.BuffGroup{}
//	for _, bgs := range e.BuffGroups {
//		buffGroups = append(buffGroups, encBattle.BuffGroup(bgs))
//	}
//
//	return &battle.S2C_TickBuff{
//		AtMs:       int32(e.AtMs),
//		BuffGroups: buffGroups,
//	}
//}
//
//func (EncBattle) PrepareSkill(e *btevt.PrepareSkill) *battle.S2C_PrepareSkill {
//	buffEvents := []*battle.BuffEvent{}
//	for _, be := range e.BuffEvents {
//		buffEvents = append(buffEvents, encBattle.BuffEvent(be))
//	}
//	return &battle.S2C_PrepareSkill{
//		FighterId:     int32(e.FighterID),
//		SkillId:       int32(e.SkillID),
//		MainTargetIds: protoIntSlice(e.MainTargetIDs),
//		AtMs:          int32(e.AtMs),
//		SingTime:      int32(e.SingTime),
//		GuideNum:      int32(e.GuideNum),
//		GuideMaxNum:   int32(e.GuideMaxNum),
//		FighterEnergy: encBattle.FighterEnergy(e.FighterEnergy),
//		BuffEvents:    buffEvents,
//	}
//}
//
//func (EncBattle) InstructionResult(e *btevt.InstructionResult) *battle.S2C_SkillResult {
//	damages := []*battle.Damage{}
//	for _, d := range e.Damages {
//		damages = append(damages, encBattle.Damage(d))
//	}
//
//	buffGroupsBefore := []*battle.BuffGroup{}
//	for _, bgb := range e.BuffGroupsBefore {
//		buffGroupsBefore = append(buffGroupsBefore, encBattle.BuffGroup(bgb))
//	}
//
//	buffGroupsAfter := []*battle.BuffGroup{}
//	for _, bga := range e.BuffGroupsAfter {
//		buffGroupsAfter = append(buffGroupsAfter, encBattle.BuffGroup(bga))
//	}
//
//	fighterEnergys := []*battle.FighterEnergy{}
//	for _, v := range e.FighterEnergys {
//		fighterEnergys = append(fighterEnergys, encBattle.FighterEnergy(v))
//	}
//
//	return &battle.S2C_SkillResult{
//		AtMs:             int32(e.AtMs),
//		Result:           resultDict[e.Result].Enum(),
//		FighterId:        int32(int(e.UnitID)),
//		SkillId:          int32(int(e.Skill)),
//		ReleaseLocation:  encBattle.Position(e.ReleasePos),
//		Damages:          damages,
//		AfterLocation:    encBattle.Position(e.AfterPos),
//		BuffGroupsBefore: buffGroupsBefore,
//		BuffGroupsAfter:  buffGroupsAfter,
//		LeechInfo:        encBattle.LeechInfo(e.LeechInfo),
//		StanceInfo:       encBattle.StanceInfo(e.StanceInfo),
//		MainTargetIds:    protoIntSlice(e.MainTargetIds),
//		Player:           encBattle.Player(&e.Player),
//		FighterEnergys:   fighterEnergys,
//	}
//}
//
//func (EncBattle) TeamMateSkill(e *btevt.TeamMateSkill) *battle.S2C_TeamMateInstruction {
//	return &battle.S2C_TeamMateInstruction{
//		FighterId: int32(e.FighterId),
//		SkillId:   int32(e.SkillId),
//	}
//}
//
//func (EncBattle) FighterSkillCdInfo(e *btevt.FighterSkillCdInfo) *battle.S2C_FighterSkillCdInfo {
//	FighterSkillCds := []*battle.FighterSkillCd{}
//	for _, v := range e.FighterSkillCds {
//		FighterSkillCds = append(FighterSkillCds, encBattle.FighterSkillCd(v))
//	}
//
//	return &battle.S2C_FighterSkillCdInfo{
//		FighterSkillCds: FighterSkillCds,
//		AtMs:            int32(e.AtMs),
//	}
//}
//
//func (EncBattle) FighterEnergyInfo(e *btevt.FighterEnergyInfo) *battle.S2C_FighterEnergyInfo {
//	fighterEnergys := []*battle.FighterEnergy{}
//	for _, v := range e.FighterEnergys {
//		fighterEnergys = append(fighterEnergys, encBattle.FighterEnergy(v))
//	}
//
//	return &battle.S2C_FighterEnergyInfo{
//		FighterEnergys: fighterEnergys,
//		AtMs:           int32(e.AtMs),
//	}
//}
//
//func (EncBattle) FighterEnergy(e btevt.FighterEnergy) *battle.FighterEnergy {
//	return &battle.FighterEnergy{
//		FighterId:      int32(e.FighterID),
//		Energy:         int32(e.Energy),
//		DeathFighterId: int32(e.DeathFighterID),
//	}
//}
//
//func (EncBattle) CancelSkill(e *btevt.CancelSkill) *battle.CancelSkill {
//	interruptEvents := []*battle.InterruptEvent{}
//	for _, v := range e.InterruptEvents {
//		interruptEvents = append(interruptEvents, encBattle.InterruptEvent(v))
//	}
//
//	return &battle.CancelSkill{
//		InterruptEvents: interruptEvents,
//		AtMs:            int32(e.AtMs),
//	}
//}
//
//func (EncBattle) FighterSkillCd(e btevt.FighterSkillCd) *battle.FighterSkillCd {
//	SkillCds := []*battle.SkillCd{}
//	for _, v := range e.SkillCds {
//		SkillCds = append(SkillCds, encBattle.SkillCd(v))
//	}
//
//	return &battle.FighterSkillCd{
//		FighterId: int32(e.FighterId),
//		SkillCds:  SkillCds,
//	}
//}
//
//func (EncBattle) SkillCd(e btevt.SkillCd) *battle.SkillCd {
//	return &battle.SkillCd{
//		SkillId:  int32(e.SkillId),
//		CdTime:   int32(e.CdTime),
//		WaitTime: int32(e.WaitTime),
//	}
//}
//
//func (EncBattle) SkillTarget(e *btevt.SkillTarget) *battle.S2C_SkillTarget {
//	fighterSkillTargetInfos := []*battle.FighterSkillTargetInfo{}
//	for _, v := range e.FighterSkillTargetInfos {
//		fighterSkillTargetInfos = append(fighterSkillTargetInfos, encBattle.FighterSkillTargetInfo(v))
//	}
//
//	return &battle.S2C_SkillTarget{
//		FighterSkillTargetInfos: fighterSkillTargetInfos,
//	}
//}
//
//func (EncBattle) FighterSkillTargetInfo(e btevt.FighterSkillTargetInfo) *battle.FighterSkillTargetInfo {
//	skillTargets := []*battle.SkillTargetInfo{}
//	for _, v := range e.SkillTargets {
//		skillTargets = append(skillTargets, encBattle.SkillTargetInfo(v))
//	}
//
//	return &battle.FighterSkillTargetInfo{
//		FighterId:        int32(e.FighterId),
//		SkillTargetInfos: skillTargets,
//	}
//}
//
//func (EncBattle) SkillTargetInfo(e btevt.SkillTargetInfo) *battle.SkillTargetInfo {
//	targetIdInfos := []*battle.TargetIdInfo{}
//	for _, v := range e.TargetIdInfos {
//		targetIdInfos = append(targetIdInfos, encBattle.TargetIdInfo(v))
//	}
//
//	return &battle.SkillTargetInfo{
//		SkillId:       int32(e.SkillId),
//		TargetIdInfos: targetIdInfos,
//		BAll:          e.BAll),
//	}
//}
//
//func (EncBattle) TargetIdInfo(e btevt.TargetIdInfo) *battle.TargetIdInfo {
//	return &battle.TargetIdInfo{
//		FirstTargetId:   int32(e.FirstTargetId),
//		SecondTargetIds: protoIntSlice(e.SecondTargetIds),
//	}
//}
//
//func (EncBattle) BuffGroup(e btevt.BuffGroup) *battle.BuffGroup {
//	buffEvents := []*battle.BuffEvent{}
//	for _, be := range e.BuffEvents {
//		buffEvents = append(buffEvents, encBattle.BuffEvent(be))
//	}
//
//	return &battle.BuffGroup{
//		BuffEvents: buffEvents,
//	}
//}
//
//func (EncBattle) BuffEvent(e btevt.BuffEvent) *battle.BuffEvent {
//	damages := []*battle.Damage{}
//	for _, d := range e.Damages {
//		damages = append(damages, encBattle.Damage(d))
//	}
//
//	fighterEnergys := []*battle.FighterEnergy{}
//	for _, v := range e.FighterEnergys {
//		fighterEnergys = append(fighterEnergys, encBattle.FighterEnergy(v))
//	}
//
//	return &battle.BuffEvent{
//		Event:          eventDict[e.Event].Enum(),
//		Buff:           encBattle.Buff(e.Buff),
//		Damages:        damages,
//		FighterHpInfo:  encBattle.FighterHpInfo(e.FighterHpInfo),
//		InterruptEvent: encBattle.InterruptEvent(e.InterruptEvent),
//		CasterId:       int32(e.CasterId),
//		FighterEnergys: fighterEnergys,
//	}
//}
//
//func (EncBattle) InterruptEvent(e btevt.InterruptEvent) *battle.InterruptEvent {
//	return &battle.InterruptEvent{
//		FighterId: int32(e.FighterID),
//		SkillId:   int32(e.SkillID),
//	}
//
//}
//
//func (EncBattle) Buff(e btevt.Buff) *battle.Buff {
//	return &battle.Buff{
//		SeqId:     int32(e.SeqId),
//		BuffId:    int32(e.BuffId),
//		FighterId: int32(e.FighterId),
//		LayerNum:  int32(e.LayerNum),
//		AddTime:   int32(e.AddTime),
//		EndTime:   int32(e.EndTime),
//	}
//}
//
//func (EncBattle) StanceInfo(e btevt.StanceInfo) *battle.StanceInfo {
//	return &battle.StanceInfo{
//		FighterId: int32(e.FighterID),
//		Stance:    int32(e.Stance),
//	}
//}
//
//func (EncBattle) LeechInfo(e btevt.LeechInfo) *battle.LeechInfo {
//	return &battle.LeechInfo{
//		Leech:         int32(e.Leech),
//		FighterHpInfo: encBattle.FighterHpInfo(e.FighterHpInfo),
//	}
//}
//
//func (EncBattle) FighterHpInfo(e btevt.FighterHpInfo) *battle.FighterHpInfo {
//	return &battle.FighterHpInfo{
//		FighterId: int32(int(e.FighterID)),
//		Hp:        int32(e.Hp),
//		MaxHp:     int32(e.MaxHp),
//		Shield:    int32(e.Shield),
//	}
//}
//
//func (EncBattle) Damage(e btevt.Damage) *battle.Damage {
//	buffEvents := []*battle.BuffEvent{}
//	for _, be := range e.BuffEvents {
//		buffEvents = append(buffEvents, encBattle.BuffEvent(be))
//	}
//
//	return &battle.Damage{
//		Result:        resultDict[e.Result].Enum(),
//		Amount:        int32(e.Point),
//		FighterId:     int32(int(e.UnitID)),
//		NewLocation:   encBattle.Position(e.AfterPos),
//		Interrupted:   &e.Interrupted,
//		Cc:            e.CC),
//		FighterHpInfo: encBattle.FighterHpInfo(e.FighterHpInfo),
//		BuffEvents:    buffEvents,
//	}
//}
//
//func (EncBattle) ChangeAutoBattle(e *btevt.ChangeAutoBattle) *battle.S2C_ChangeAutoBattle {
//	return &battle.S2C_ChangeAutoBattle{
//		FighterId: int32(e.FighterID),
//		AtMs:      int32(e.AtMs),
//	}
//}
//
//func (EncBattle) ChooseSuperSkill(e *btevt.ChooseSuperSkill) *battle.S2C_ChooseSuperSkill {
//	return &battle.S2C_ChooseSuperSkill{
//		FighterId: int32(e.FighterID),
//		SkillId:   int32(e.SkillID),
//		AtMs:      int32(e.AtMs),
//	}
//}
//
//func (EncBattle) SAutoBattle(e *btevt.SAutoBattle) *battle.S2C_AutoBattle {
//	return &battle.S2C_AutoBattle{
//		BAuto: e.BAuto),
//		AtMs:  int32(e.AtMs),
//	}
//}
//
//func (EncBattle) Position(p btevt.Position) *battle.Location {
//	return &battle.Location{X: int32(p.X), Y: int32(p.Y)}
//}
//
//func (EncBattle) PhaseEnd(e *btevt.PhaseEnd) *battle.S2C_PhaseEnd {
//	phase := battle.S2C_PhaseEnd_Phase1
//	if e.Phase == btevt.DogFightPhase {
//		phase = battle.S2C_PhaseEnd_Phase2
//	}
//	return &battle.S2C_PhaseEnd{Phase: phase.Enum(), State: encBattle.State(e.State)}
//}
//
//func (EncBattle) GameOver(e *btevt.GameOver) *battle.S2C_GameOver {
//	ret := battle.S2C_GameOver_right
//	if e.Ret == btevt.Right {
//		ret = battle.S2C_GameOver_right
//	} else if e.Ret == btevt.Left {
//		ret = battle.S2C_GameOver_left
//	} else if e.Ret == btevt.Draw {
//		ret = battle.S2C_GameOver_draw
//	}
//	return &battle.S2C_GameOver{
//		Ret: &ret,
//	}
//}

func (EncGame) ResNPCPickPrepare(e *gmevt.ResNPCPickPrepare) *field.S2C_PickPrepare {
	if e == nil {
		return nil
	}
	return &field.S2C_PickPrepare{
		Ok:  e.Ok,
		Num: int32(e.Num),
		Max: int32(e.Max),
	}
}

func (enc EncGame) ResNPCPick(e *gmevt.ResNPCPick) *field.S2C_PickWith {
	if e == nil {
		return nil
	}
	return &field.S2C_PickWith{
		Ok:    e.Ok,
		Loots: enc.Loots(e.Loots),
		Num:   int32(e.Num),
		Max:   int32(e.Max),
	}
}

func (EncGame) ResNPCFishingPrepare(e *gmevt.ResNPCFishingPrepare) *field.S2C_FishingPrepare {
	if e == nil {
		return nil
	}
	return &field.S2C_FishingPrepare{
		Ok:  e.Ok,
		Num: int32(e.Num),
		Max: int32(e.Max),
	}
}

func (enc EncGame) ResNPCFishing(e *gmevt.ResNPCFishing) *field.S2C_FishingWith {
	if e == nil {
		return nil
	}
	return &field.S2C_FishingWith{
		Ok:    e.Ok,
		Loots: enc.Loots(e.Loots),
		Num:   int32(e.Num),
		Max:   int32(e.Max),
	}
}

func (EncGame) ResNPCRoastPrepare(e *gmevt.ResNPCRoastPrepare) *field.S2C_RoastPrepare {
	if e == nil {
		return nil
	}
	return &field.S2C_RoastPrepare{
		Ok: e.Ok,
	}
}

func (enc EncGame) ResNPCRoast(e *gmevt.ResNPCRoast) *field.S2C_RoastWith {
	if e == nil {
		return nil
	}
	return &field.S2C_RoastWith{
		Ok:    e.Ok,
		Loots: enc.Loots(e.Loots),
	}
}

func (enc EncGame) ResUnlockRace(e *gmevt.ResUnlockRace) *action.S2C_UnlockRace {
	return &action.S2C_UnlockRace{}
}

func (enc EncGame) FameDonate(e *gmevt.FameDonate) *shared.FameDonate {
	return &shared.FameDonate{
		Id:         int32(e.ID),
		Num:        int32(e.Num),
		PresentId:  int32(e.PresentID),
		PresentNum: int32(e.PresentNum),
		Exp:        int32(e.Exp),
		LeftNum:    int32(e.LeftNum),
	}
}

func (enc EncGame) FameDonateSlice(e []gmevt.FameDonate) []*shared.FameDonate {
	rv := []*shared.FameDonate{}
	for _, v := range e {
		rv = append(rv, enc.FameDonate(&v))
	}
	return rv
}

func (enc EncGame) ResDonateFame(e *gmevt.ResDonateFame) *action.S2C_DonateFame {
	return &action.S2C_DonateFame{
		FameDonate: enc.FameDonate(e.FameDonate),
	}
}

func (enc EncGame) GoodsCond(conds []gmevt.GoodsCond) []*shared.Goods_Condition {
	rv := []*shared.Goods_Condition{}

	for _, v := range conds {
		condType := shared.Goods_CondType(v.CondType)
		rv = append(rv, &shared.Goods_Condition{
			CondType:  condType,
			CondID:    int32(v.CondID),
			CondValue: int32(v.CondValue),
		})
	}

	return rv
}

func (enc EncGame) Goods(e *gmevt.Goods) *shared.Goods {
	currency := shared.Loot_LootType(e.CostType)
	numType := shared.Goods_NumType(e.NumType)
	itemType := shared.Loot_LootType(e.ItemType)
	return &shared.Goods{
		Id:       int32(e.ID),
		ItemType: itemType,
		ItemId:   int32(e.ItemID),
		Conds:    enc.GoodsCond(e.Conds),
		Currency: currency,
		Price:    int32(e.CostNum),
		NumType:  numType,
		LimitNum: int32(e.LimitNum),
		LeftNum:  int32(e.LeftNum),
		Cost:     enc.Loot(e.Cost),
	}
}

func (enc EncGame) GoodsSlice(e []gmevt.Goods) []*shared.Goods {
	rv := []*shared.Goods{}
	for _, v := range e {
		rv = append(rv, enc.Goods(&v))
	}
	return rv
}

func (enc EncGame) ResBuyGoods(e *gmevt.ResBuyGoods) *action.S2C_BuyGoods {
	return &action.S2C_BuyGoods{
		Goods: enc.Goods(e.Goods),
	}
}

func (EncGame) ResQueryFameDonate(e *gmevt.ResQueryFameDonate) *query.S2C_QueryFameDonate {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryFameDonate{
		FameDonate: encGame.FameDonateSlice(e.FameDonates),
	}
}

func (EncGame) ResQueryGoods(e *gmevt.ResQueryGoods) *query.S2C_QueryGoods {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryGoods{
		Goods: encGame.GoodsSlice(e.Goods),
	}
}

func (EncGame) ResQueryServerTime(e *gmevt.ResQueryServerTime) *query.S2C_QueryServerTime {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryServerTime{
		Timestamp: e.ServerTime,
	}
}

func (enc EncGame) Statistic(e *gmevt.Statistic) *shared.Statistic {
	return &shared.Statistic{
		Tid: int32(e.UID),
		Num: int32(e.Num),
	}
}

func (enc EncGame) StatisticSlice(e []gmevt.Statistic) []*shared.Statistic {
	rv := []*shared.Statistic{}
	for _, v := range e {
		rv = append(rv, enc.Statistic(&v))
	}
	return rv
}

func (EncGame) ResQueryStatistics(e *gmevt.ResQueryStatistics) *query.S2C_QueryStatistics {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryStatistics{
		Statistics: encGame.StatisticSlice(e.Statistics),
	}
}

func (EncGame) ResQueryAchievements(e *gmevt.ResQueryAchievements) *query.S2C_QueryAchievements {
	if e == nil {
		return nil
	}
	return &query.S2C_QueryAchievements{
		NType:        int32(e.AchieveType),
		Achievements: encGame.Achievements(e.Achievements),
	}
}

func activityStatus(as gmevt.ActivityStatus) *query.ActivityStatus {
	return &query.ActivityStatus{
		ActivityId:     int32(as.ID),
		Activated:      as.Activated,
		StartTimestamp: as.Start.Unix(),
		EndTimestamp:   as.End.Unix(),
	}
}

func (EncGame) ResQueryActivities(e *gmevt.ResQueryActivities) *query.S2C_QueryActivities {
	var activities []*query.ActivityStatus
	for _, a := range e.Activities {
		activities = append(activities, activityStatus(a))
	}
	return &query.S2C_QueryActivities{
		Activities: activities,
	}
}

func (EncGame) ResQueryUID(e *gmevt.ResQueryUID) *query.S2C_QueryUID {
	return &query.S2C_QueryUID{
		Uid: int64(e.UID),
	}
}

func (enc EncGame) ResQueryUserDetail(e *gmevt.ResQueryUserDetail) *query.S2C_QueryUserDetail {
	return &query.S2C_QueryUserDetail{
		Ok:     e.OK,
		Result: enc.Operation(e.Result),
	}
}

func (EncGame) ResClaimQuestReward(e *gmevt.ResClaimQuestReward) *action.S2C_ClaimQuestReward {
	return &action.S2C_ClaimQuestReward{
		Ok: e.OK,
	}
}

func (EncGame) ResPKWith(e *gmevt.ResPKWith) *action.S2C_PKWith {
	return &action.S2C_PKWith{
		Ok: e.OK,
	}
}

func (EncGame) ResCircle(e *gmevt.ResCircle) *action.S2C_AssignCircle {
	return &action.S2C_AssignCircle{
		Ok: e.OK,
	}
}

func (enc EncGame) ResExtendBag(e *gmevt.ResExtendBag) *action.S2C_ExtendBag {
	return &action.S2C_ExtendBag{}
}

//*****************equipconfig begin
/*
func (enc EncGame) ResRecast(m *gmevt.ResRecast) *action.S2C_Recast {
	if m == nil {
		return nil
	}
	return &action.S2C_Recast{
		Ok:    m.OK),
		Equip: enc.Equipment(&m.Equip),
	}
}
*/
//*****************equipconfig begin

func (enc EncGame) ResIdentify(e *gmevt.ResIdentify) *action.S2C_Identify {
	if e == nil {
		return nil
	}
	return &action.S2C_Identify{
		Ok:    e.OK,
		Loots: enc.Loots(e.Loots),
	}
}

func (enc EncGame) ResComposite(e *gmevt.ResComposite) *action.S2C_Composite {
	if e == nil {
		return nil
	}
	return &action.S2C_Composite{
		Ok: e.OK,
	}
}

func (enc EncGame) ResStudyPaper(e *gmevt.ResStudyPaper) *action.S2C_StudyPaper {
	if e == nil {
		return nil
	}
	return &action.S2C_StudyPaper{
		Ok: e.OK,
	}
}

func (enc EncGame) ResTakeAchievementAward(e *gmevt.ResTakeAchievementAward) *action.S2C_TakeAchievementAward {
	if e == nil {
		return nil
	}
	return &action.S2C_TakeAchievementAward{
		Ok: e.OK,
	}
}

func (enc EncGame) ResChangeUserName(e *gmevt.ResChangeUserName) *action.S2C_ChangeUserName {
	if e == nil {
		return nil
	}
	return &action.S2C_ChangeUserName{
		Ok: e.OK,
	}
}

func protoIntSlice(s []int) []int32 {
	return intutil.IntToInt32Slice(s)
}

func (enc EncGame) ResUpgradeStove(e *gmevt.ResUpgradeStove) *action.S2C_UpgradeStove {
	if e == nil {
		return nil
	}
	return &action.S2C_UpgradeStove{
		Ok: e.OK,
	}
}

func (enc EncGame) ResUserGetExp(e *gmevt.ResUserGetExp) *action.S2C_UserGetExp {
	if e == nil {
		return nil
	}
	return &action.S2C_UserGetExp{
		Exp: int32(e.Exp),
	}
}

func (enc EncGame) ResUserGetFameExp(e *gmevt.ResUserGetFameExp) *action.S2C_UserGetFameExp {
	if e == nil {
		return nil
	}
	return &action.S2C_UserGetFameExp{
		FameId: int32(e.FameID),
		Exp:    int32(e.Exp),
	}
}

func (enc EncGame) ResUnlockMaskPart(e *gmevt.ResUnlockMaskPart) *action.S2C_UnlockMaskPart {
	if e == nil {
		return nil
	}
	return &action.S2C_UnlockMaskPart{}
}

func (enc EncGame) ResUnlockMask(e *gmevt.ResUnlockMask) *action.S2C_UnlockMask {
	if e == nil {
		return nil
	}
	return &action.S2C_UnlockMask{}
}

func (enc EncGame) ResStartEscort(e *gmevt.ResStartEscort) *action.S2C_StartEscort {
	if e == nil {
		return nil
	}
	return &action.S2C_StartEscort{
		Ok: e.OK,
	}
}

func (enc EncGame) EscortUpdate(e *gmevt.EscortUpdate) *action.S2C_EscortUpdate {
	if e == nil {
		return nil
	}
	return &action.S2C_EscortUpdate{
		RangeWarning: e.RangeWarning,
		NpcUid:       e.NPCUID,
		NpcId:        int32(e.NPCID),
	}
}

func (enc EncGame) EscortEnded(e *gmevt.EscortEnded) *action.S2C_EscortEnded {
	if e == nil {
		return nil
	}
	return &action.S2C_EscortEnded{
		Succeed: e.Succeed,
		NpcUid:  e.NPCUID,
		NpcId:   int32(e.NPCID),
	}
}

func (enc EncGame) ResStartFollow(e *gmevt.ResStartFollow) *action.S2C_Follow {
	if e == nil {
		return nil
	}
	return &action.S2C_Follow{
		Following: e.Following,
		NpcUid:    e.NPCUID,
	}
}

func (enc EncGame) ResStopFollow(e *gmevt.ResStopFollow) *action.S2C_Follow {
	if e == nil {
		return nil
	}
	return &action.S2C_Follow{
		Following: e.Following,
		NpcUid:    e.NPCUID,
	}
}

func (enc EncGame) ResAdWatched(e *gmevt.ResAdWatched) *action.S2C_AdWatched {
	if e == nil {
		return nil
	}
	return &action.S2C_AdWatched{
		HasNext:   e.HasNext,
		RemainNum: int32(e.RemainNum),
	}
}

func (enc EncGame) ResLoginReward(e *gmevt.ResLoginReward) *action.S2C_LoginReward {
	if e == nil {
		return nil
	}
	return &action.S2C_LoginReward{
		CurrID:   int32(e.CurrID),
		CurrIdx:  int32(e.CurrIdx),
		IsGot:    e.IsGot,
		TotalGot: int32(e.TotalGot),
		PrevID:   int32(e.PrevID),
		NextID:   int32(e.NextID),
	}
}

func (enc EncGame) ResGainLoginReward(e *gmevt.ResGainLoginReward) *action.S2C_GainLoginReward {
	if e == nil {
		return nil
	}
	return &action.S2C_GainLoginReward{
		CurrID:  int32(e.CurrID),
		CurrIdx: int32(e.CurrIdx),
		Ok:      e.Ok,
	}
}

func (enc EncGame) ResSocialShared(e *gmevt.ResSocialShared) *action.S2C_SocialShared {
	if e == nil {
		return nil
	}
	return &action.S2C_SocialShared{
		Ok: e.Ok,
	}
}

func (enc EncGame) ResAdInfo(e *gmevt.ResAdInfo) *action.S2C_AdInfo {
	if e == nil {
		return nil
	}
	return &action.S2C_AdInfo{
		HasNext:   e.HasNext,
		RemainNum: int32(e.RemainNum),
	}
}

// func (enc EncGame) ResActiveRewardList(e *gmevt.ResActiveRewardList) *action.S2C_ActiveRewardList {
// 	if e == nil {
// 		return nil
// 	}
// 	return &action.S2C_ActiveRewardList{
// 		Point:   int32(e.Point),
// 		Entries: enc.ActiveRewardEntries(e.Entries),
// 	}
// }

// func (enc EncGame) ResGainActiveReward(e *gmevt.ResGainActiveReward) *action.S2C_GainActiveReward {
// 	if e == nil {
// 		return nil
// 	}
// 	return &action.S2C_GainActiveReward{
// 		ActiveRewardID: int32(e.ActiveRewardID),
// 		Ok:             e.Ok),
// 	}
// }

// func (enc EncGame) ActiveRewardEntries(ls []*gmevt.ActiveRewardEntry) []*action.ActiveRewardEntry {
// 	rv := []*action.ActiveRewardEntry{}
// 	for _, v := range ls {
// 		rv = append(rv, enc.ActiveRewardEntry(v))
// 	}

// 	return rv
// }

// func (enc EncGame) ActiveRewardEntry(v *gmevt.ActiveRewardEntry) *action.ActiveRewardEntry {
// 	return &action.ActiveRewardEntry{
// 		ActiveRewardID: int32(v.ActiveRewardID),
// 		GainStatus:     int32(v.GainStatus),
// 		Loots:          enc.Loots(v.Loots),
// 	}
// }

func (enc EncGame) ResAskForQuest(e *gmevt.ResAskForQuest) *action.S2C_AskForQuest {
	if e == nil {
		return nil
	}
	return &action.S2C_AskForQuest{
		QuestId: int32(e.QuestID),
	}
}
