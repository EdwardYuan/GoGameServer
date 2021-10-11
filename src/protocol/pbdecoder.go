//NOTE: the mechanism in this file is kind of a hack, please BE AWARE!
package message

import (
	"errors"
	"hope_server/event"
	"hope_server/kit/logz"
	"pb/base"
	"reflect"
	"runtime/debug"

	"github.com/golang/protobuf/proto"
	/*
		"hope_server/config"
		"hope_server/event"
		"hope_server/game/gmevt"
		"hope_server/intutil"
		"hope_server/kit/logz"

		"pb/action"
		"pb/base"
		"pb/chat"
		"pb/field"
		"pb/friend"
		"pb/query"
	) */)

var ErrGenericDecodeError = errors.New("generic decode error")

func ParseProtobufEvent(bs []byte) (seq int, e event.Event, typeStr string, headerLen, bodyLen int, err error) {
	if !config.FlagDev {
		defer func() {
			if r := recover(); r != nil {
				logz.Error("recover from proto event parse error", "event_name", event.Name(e),
					"recover_result", r, "debug_stack", string(debug.Stack()))
				switch r := r.(type) {
				case string:
					err = errors.New(r)
				case error:
					err = r
				default:
					err = ErrGenericDecodeError
				}
			}
		}()
	}

	m, headerLen, bodyLen, err := Decode(bs)
	if err != nil {
		return
	}
	seq = int(m.ID)
	typeStr = m.Type()

	methodValue, ok := decMethods[typeStr]
	if ok {
		// 手动从protobuf转到event
		rv := methodValue.Call([]reflect.Value{reflect.ValueOf(m.Body)})
		e = rv[0].Interface()
	} else {
		// 直接解码
		e = m.Body.(event.Event)
	}

	return
}

type PBDecoder interface {
	PkgName() string
}

//type DecBattle struct{}
type DecChat struct{}
type DecQuery struct{}
type DecLogin struct{}
type DecShared struct{}
type DecField struct{}
type DecAction struct{}
type DecFriend struct{}

//func (d DecBattle) PkgName() string {
//	return "battle"
//}
func (d DecChat) PkgName() string {
	return "chat"
}
func (d DecQuery) PkgName() string {
	return "query"
}
func (d DecLogin) PkgName() string {
	return "login"
}
func (d DecShared) PkgName() string {
	return "shared"
}
func (d DecField) PkgName() string {
	return "field"
}
func (d DecAction) PkgName() string {
	return "action"
}
func (d DecFriend) PkgName() string {
	return "friend"
}

var (
	//	decBattle DecBattle
	decChat   DecChat
	decQuery  DecQuery
	decLogin  DecLogin
	decShared DecShared
	decField  DecField
	decAction DecAction
	decFriend DecFriend
)

var decMethods map[string]reflect.Value = map[string]reflect.Value{}

func init() {
	//	registerDecoder(decBattle)
	registerDecoder(decChat)
	registerDecoder(decQuery)
	registerDecoder(decLogin)
	registerDecoder(decShared)
	registerDecoder(decField)
	registerDecoder(decAction)
	registerDecoder(decFriend)
}

func registerDecoder(dec PBDecoder) {
	v := reflect.ValueOf(dec)
	t := v.Type()
	n := t.NumMethod()
	for i := 0; i < n; i++ {
		method := t.Method(i)
		if method.Name == "PkgName" { // this is in interface{}
			continue
		}
		name := dec.PkgName() + "." + method.Name
		if _, ok := decMethods[name]; ok {
			panic("decode method " + name + " already registered")
		}
		decMethods[name] = v.Method(i)
	}
}

func (DecChat) CSQ_ChatMessage(pm proto.Message) *gmevt.CsqChatMessage {
	m := pm.(*chat.CSQ_ChatMessage)
	return &gmevt.CsqChatMessage{
		ChatId:      m.GetChatId(),
		ChatType:    m.GetChatType(),
		Msg:         m.GetMsg(),
		RoleId:      m.GetRoleId(),
		AudioStream: m.GetAudioStream(),
	}
}

func (DecQuery) C2S_QueryUser(pm proto.Message) *gmevt.ReqQueryUser {
	// m := pm.(*query.C2S_QueryUser)
	return &gmevt.ReqQueryUser{
		// UserVersion: decShared.UserVersion(m.GetUserVersion()),
	}
}

func (DecQuery) C2S_QueryOtherUser(pm proto.Message) *gmevt.ReqQueryOtherUser {
	m := pm.(*query.C2S_QueryOtherUser)

	keys := []gmevt.QueryOtherUserKey{}
	for _, v := range m.Keys {
		keys = append(keys, gmevt.QueryOtherUserKey(v))
	}
	return &gmevt.ReqQueryOtherUser{
		Who:  int(m.GetWho()),
		Keys: keys,
		Type: int(m.GetType()),
	}
}

func (DecQuery) C2S_QueryFameDonate(pm proto.Message) *gmevt.ReqQueryFameDonate {
	m := pm.(*query.C2S_QueryFameDonate)
	return &gmevt.ReqQueryFameDonate{
		FameID:  int(m.GetFameId()),
		Version: int(m.GetVersion()),
	}
}

func (DecQuery) C2S_QueryGoods(pm proto.Message) *gmevt.ReqQueryGoods {
	m := pm.(*query.C2S_QueryGoods)
	return &gmevt.ReqQueryGoods{
		ShopID:  int(m.GetShopId()),
		Version: int(m.GetVersion()),
	}
}

func (DecQuery) C2S_QueryServerTime(pm proto.Message) *gmevt.ReqQueryServerTime {
	return &gmevt.ReqQueryServerTime{}
}

func (DecQuery) C2S_QueryStatistics(pm proto.Message) *gmevt.ReqQueryStatistics {
	return &gmevt.ReqQueryStatistics{}
}

func (DecQuery) C2S_QueryAchievements(pm proto.Message) *gmevt.ReqQueryAchievements {
	m := pm.(*query.C2S_QueryAchievements)
	return &gmevt.ReqQueryAchievements{
		Version:     int(m.GetVersion()),
		AchieveType: int(m.GetNType()),
	}
}

func (DecQuery) C2S_QueryActivities(pm proto.Message) *gmevt.ReqQueryActivities {
	m := pm.(*query.C2S_QueryActivities)
	return &gmevt.ReqQueryActivities{
		ActivityIDs: intutil.Int32ToIntSlice(m.GetActivityIds()),
	}
}

func (DecQuery) C2S_QueryUID(pm proto.Message) *gmevt.ReqQueryUID {
	return &gmevt.ReqQueryUID{}
}

func (DecQuery) C2S_QueryUserDetail(pm proto.Message) *gmevt.ReqQueryUserDetail {
	m := pm.(*query.C2S_QueryUserDetail)
	return &gmevt.ReqQueryUserDetail{
		Who:  int(m.GetWho()),
		Item: gmevt.QueryUserDetailItem(m.GetWhat()),
		Uid:  m.GetUid(),
	}
}

// func (DecShared) UserVersion(m *shared.UserVersion) *gmevt.UserVersion {
// 	return &gmevt.UserVersion{}
// }

func (DecShared) Ping(pm proto.Message) *gmevt.Ping {
	return new(gmevt.Ping)
}

func (DecField) C2S_EnterField(pm proto.Message) *gmevt.ReqEnterField {
	m := pm.(*field.C2S_EnterField)
	pos := 0
	if m.GetPos() != nil {
		pos = int(m.GetPos().GetX())
	}
	return &gmevt.ReqEnterField{
		FieldID: int(m.GetField().GetId()),
		Pos:     pos,
	}
}

func (DecField) C2S_EnterHome(pm proto.Message) *gmevt.ReqEnterHome {
	m := pm.(*field.C2S_EnterHome)
	return &gmevt.ReqEnterHome{
		UID:           int(m.GetPlayerUid()),
		HearthStoneID: int(m.GetHearthStoneID()),
	}
}

func (DecField) C2S_Update(pm proto.Message) *gmevt.CliFieldUpdate {
	m := pm.(*field.C2S_Update)
	return &gmevt.CliFieldUpdate{
		Pos: int(m.GetPos().GetX()),
	}
}

func (DecField) C2S_TalkWith(pm proto.Message) *gmevt.NPCTalk {
	m := pm.(*field.C2S_TalkWith)
	return &gmevt.NPCTalk{
		NPCUID:    m.GetNpcUid(),
		NPCID:     int(m.GetNpcId()),
		ContentID: int(m.GetContentId()),
		ForQuest:  int(m.GetForQuest()),
	}
}

func (DecField) C2S_BattleWith(pm proto.Message) *gmevt.NPCBattle {
	m := pm.(*field.C2S_BattleWith)
	return &gmevt.NPCBattle{
		NPCUID:     m.GetNpcUid(),
		NPCID:      int(m.GetNpcId()),
		Difficulty: int(m.GetDifficulty()),
		StandPos:   gmevt.StandPosition(m.GetStandPos()),
	}
}

func (DecField) C2S_PickPrepare(pm proto.Message) *gmevt.NPCPickPrepare {
	m := pm.(*field.C2S_PickPrepare)
	return &gmevt.NPCPickPrepare{
		NPCUID: m.GetNpcUid(),
		NPCID:  int(m.GetNpcId()),
	}
}

func (DecField) C2S_PickWith(pm proto.Message) *gmevt.NPCPick {
	m := pm.(*field.C2S_PickWith)
	return &gmevt.NPCPick{
		NPCUID: m.GetNpcUid(),
		NPCID:  int(m.GetNpcId()),
		ItemID: int(m.GetItemId()),
	}
}

func (DecField) C2S_FishingPrepare(pm proto.Message) *gmevt.NPCFishingPrepare {
	m := pm.(*field.C2S_FishingPrepare)
	return &gmevt.NPCFishingPrepare{
		NPCUID: m.GetNpcUid(),
		NPCID:  int(m.GetNpcId()),
	}
}

func (DecField) C2S_FishingWith(pm proto.Message) *gmevt.NPCFishing {
	m := pm.(*field.C2S_FishingWith)
	return &gmevt.NPCFishing{
		NPCUID: m.GetNpcUid(),
		NPCID:  int(m.GetNpcId()),
		ItemID: int(m.GetItemId()),
	}
}

func (DecField) C2S_RoastPrepare(pm proto.Message) *gmevt.NPCRoastPrepare {
	m := pm.(*field.C2S_RoastPrepare)
	return &gmevt.NPCRoastPrepare{
		NPCUID: m.GetNpcUid(),
		NPCID:  int(m.GetNpcId()),
	}
}

func (DecField) C2S_RoastWith(pm proto.Message) *gmevt.NPCRoast {
	m := pm.(*field.C2S_RoastWith)
	return &gmevt.NPCRoast{
		NPCUID:    m.GetNpcUid(),
		NPCID:     int(m.GetNpcId()),
		FormulaID: int(m.GetItemId()),
	}
}

// action
func (DecAction) C2S_UpdateFormation(pm proto.Message) *gmevt.ReqUpdateFormation {
	m := pm.(*action.C2S_UpdateFormation)
	var form gmevt.Formation
	for _, v := range m.GetForm().GetUids() {
		form.MercUIDs = append(form.MercUIDs, v)
	}
	return &gmevt.ReqUpdateFormation{
		Form: form,
	}
}

func (DecAction) C2S_SortItem(pm proto.Message) *gmevt.ReqSortItem {
	return &gmevt.ReqSortItem{}
}

func (DecAction) C2S_SellItem(pm proto.Message) *gmevt.ReqSellItem {
	m := pm.(*action.C2S_SellItem)
	return &gmevt.ReqSellItem{
		UID:    m.GetUid(),
		Amount: intutil.Int32ToIntSlice(m.GetAmount()),
	}
}

func (DecAction) C2S_UseItem(pm proto.Message) *gmevt.ReqUseItem {
	m := pm.(*action.C2S_UseItem)
	return &gmevt.ReqUseItem{
		UID:    m.GetUid(),
		Amount: int(m.GetAmount()),
	}
}

func (DecAction) C2S_EquipMask(pm proto.Message) *gmevt.ReqEquipMask {
	m := pm.(*action.C2S_EquipMask)
	return &gmevt.ReqEquipMask{
		UID: m.GetUid(),
	}
}

func (DecAction) C2S_EquipEquipment(pm proto.Message) *gmevt.ReqEquipEquipment {
	m := pm.(*action.C2S_EquipEquipment)
	return &gmevt.ReqEquipEquipment{
		UID: m.GetUid(),
	}
}

func (DecAction) C2S_UnEquipEquipment(pm proto.Message) *gmevt.ReqUnEquipEquipment {
	m := pm.(*action.C2S_UnEquipEquipment)
	return &gmevt.ReqUnEquipEquipment{
		UID: m.GetUid(),
	}
}

/*
func (DecAction) C2S_EquipMount(pm proto.Message) *gmevt.ReqEquipMount {
	m := pm.(*action.C2S_EquipMount)
	return &gmevt.ReqEquipMount{
		UID: m.GetUid(),
	}
}

func (DecAction) C2S_UnEquipMount(pm proto.Message) *gmevt.ReqUnEquipMount {
	m := pm.(*action.C2S_UnEquipMount)
	return &gmevt.ReqUnEquipMount{
		UID: m.GetUid(),
	}
}
*/

func (DecAction) C2S_EquipPartner(pm proto.Message) *gmevt.ReqEquipPartner {
	m := pm.(*action.C2S_EquipPartner)
	return &gmevt.ReqEquipPartner{
		UID: m.GetUid(),
	}
}

func (DecAction) C2S_UnEquipPartner(pm proto.Message) *gmevt.ReqUnEquipPartner {
	m := pm.(*action.C2S_UnEquipPartner)
	return &gmevt.ReqUnEquipPartner{
		UID: m.GetUid(),
	}
}

func (DecAction) C2S_UnlockRace(pm proto.Message) *gmevt.ReqUnlockRace {
	m := pm.(*action.C2S_UnlockRace)
	return &gmevt.ReqUnlockRace{
		UID:        int(m.GetTid()),
		UnlockType: int(m.GetUnlocktype()),
	}
}

func (DecAction) C2S_DonateFame(pm proto.Message) *gmevt.ReqDonateFame {
	m := pm.(*action.C2S_DonateFame)
	return &gmevt.ReqDonateFame{
		DonateID: int(m.GetDonateId()),
		Num:      int(m.GetNum()),
	}
}

func (DecAction) C2S_BuyGoods(pm proto.Message) *gmevt.ReqBuyGoods {
	m := pm.(*action.C2S_BuyGoods)
	return &gmevt.ReqBuyGoods{
		NPCUID:    m.GetNpcUid(),
		NPCID:     int(m.GetNpcId()),
		GoodsID:   int(m.GetGoodsId()),
		Num:       int(m.GetNum()),
		CouponUID: m.GetCouponUID(),
	}
}

func (DecAction) C2S_ClaimQuestReward(pm proto.Message) *gmevt.ReqClaimQuestReward {
	m := pm.(*action.C2S_ClaimQuestReward)
	return &gmevt.ReqClaimQuestReward{
		QuestType: config.QuestType(m.GetQuestType()),
		QuestID:   int(m.GetQuestId()),
	}
}

// func (DecAction) C2S_PKWith(pm proto.Message) *gmevt.ReqPKWith {
// 	m := pm.(*action.C2S_PKWith)
// 	return &gmevt.ReqPKWith{
// 		UID: int(m.GetPlayerUid()),
// 	}
// }

func (DecAction) C2S_AssignCircle(pm proto.Message) *gmevt.ReqCircle {
	m := pm.(*action.C2S_AssignCircle)
	return &gmevt.ReqCircle{
		NPCUID:  m.GetNpcUid(),
		NPCID:   int(m.GetNpcId()),
		Publish: bool(m.GetPublish()),
	}
}

func (DecAction) C2S_ExtendBag(pm proto.Message) *gmevt.ReqExtendBag {
	m := pm.(*action.C2S_ExtendBag)
	return &gmevt.ReqExtendBag{
		ExtType: int(m.GetExtType()),
	}
}

//func (DecAction) C2S_DeleteMail(pm proto.Message) *gmevt.ReqDeleteMail {
//	m := pm.(*action.C2S_DeleteMail)
//	return &gmevt.ReqDeleteMail{
//		UIDs: m.Uids,
//	}
//}
//
//func (DecAction) C2S_TakeMailItems(pm proto.Message) *gmevt.ReqTakeMailItems {
//	m := pm.(*action.C2S_TakeMailItems)
//	return &gmevt.ReqTakeMailItems{
//		UIDs: m.Uids,
//	}
//}

func (DecAction) C2S_TakeAchievementAward(pm proto.Message) *gmevt.ReqTakeAchievementAward {
	m := pm.(*action.C2S_TakeAchievementAward)
	return &gmevt.ReqTakeAchievementAward{
		UID: int(m.GetUid()),
	}
}

func (DecAction) C2S_AdWatched(pm proto.Message) *gmevt.ReqAdWatched {
	return &gmevt.ReqAdWatched{}
}

func (DecAction) C2S_AdInfo(pm proto.Message) *gmevt.ReqAdInfo {
	return &gmevt.ReqAdInfo{}
}

func (DecAction) C2S_LoginReward(pm proto.Message) *gmevt.ReqLoginReward {
	return &gmevt.ReqLoginReward{}
}

func (DecAction) C2S_GainLoginReward(pm proto.Message) *gmevt.ReqGainLoginReward {
	m := pm.(*action.C2S_GainLoginReward)
	return &gmevt.ReqGainLoginReward{
		CurrID:  int(m.GetCurrID()),
		CurrIdx: int(m.GetCurrIdx()),
	}
}

func (DecAction) C2S_AskForQuest(pm proto.Message) *gmevt.ReqAskForQuest {
	m := pm.(*action.C2S_AskForQuest)
	return &gmevt.ReqAskForQuest{
		QuestType: config.QuestType(m.GetQuestType()),
		NPCUID:    m.GetNpcUid(),
		NPCID:     int(m.GetNpcId()),
		QuestID:   int(m.GetQuestID()),
	}
}

func (DecAction) C2S_SocialShared(pm proto.Message) *gmevt.ReqSocialShared {
	m := pm.(*action.C2S_SocialShared)
	return &gmevt.ReqSocialShared{
		Type: int(m.GetType()),
	}
}

// func (DecAction) C2S_ActiveRewardList(pm proto.Message) *gmevt.ReqActiveRewardList {
// 	return &gmevt.ReqActiveRewardList{}
// }

// func (DecAction) C2S_GainActiveReward(pm proto.Message) *gmevt.ReqGainActiveReward {
// 	m := pm.(*action.C2S_GainActiveReward)
// 	return &gmevt.ReqGainActiveReward{
// 		ActiveRewardID: int(m.GetActiveRewardID()),
// 	}
// }

// friend
func (DecFriend) C2S_RecommendFriend(pm proto.Message) *gmevt.ReqRecommendFriend {
	return &gmevt.ReqRecommendFriend{}
}

func (DecFriend) C2S_SearchNearbyUser(pm proto.Message) *gmevt.ReqSearchNearbyUser {
	return &gmevt.ReqSearchNearbyUser{}
}

func (DecFriend) C2S_LocateUser(pm proto.Message) *gmevt.ReqLocateUser {
	m := pm.(*friend.C2S_LocateUser)
	return &gmevt.ReqLocateUser{
		Keyword: m.GetKeyword(),
	}
}

func (DecFriend) C2S_AddFriendByUID(pm proto.Message) *gmevt.ReqAddFriendByUID {
	m := pm.(*friend.C2S_AddFriendByUID)
	return &gmevt.ReqAddFriendByUID{
		UID: int(m.GetUid()),
	}
}

func (DecFriend) C2S_AddFriendByName(pm proto.Message) *gmevt.ReqAddFriendByName {
	m := pm.(*friend.C2S_AddFriendByName)
	return &gmevt.ReqAddFriendByName{
		Name: m.GetName(),
	}
}

func (DecFriend) C2S_RemoveFriendByUID(pm proto.Message) *gmevt.ReqRemoveFriendByUID {
	m := pm.(*friend.C2S_RemoveFriendByUID)
	return &gmevt.ReqRemoveFriendByUID{
		UID: int(m.GetUid()),
	}
}

func (DecFriend) C2S_AddBlacklist(pm proto.Message) *gmevt.ReqAddBlacklist {
	m := pm.(*friend.C2S_AddBlacklist)
	return &gmevt.ReqAddBlacklist{
		Keyword: m.GetKeyword(),
	}
}

func (DecFriend) C2S_RemoveBlacklist(pm proto.Message) *gmevt.ReqRemoveBlacklist {
	m := pm.(*friend.C2S_RemoveBlacklist)
	return &gmevt.ReqRemoveBlacklist{
		Keyword: m.GetKeyword(),
	}
}

func (DecFriend) C2S_FriendStatusList(pm proto.Message) *gmevt.ReqFriendStatusList {
	return &gmevt.ReqFriendStatusList{}
}

func (DecFriend) C2S_RecentTeamMemberList(pm proto.Message) *gmevt.ReqRecentTeamMemberList {
	return &gmevt.ReqRecentTeamMemberList{}
}

func UserBaseInfoProtoToEvent(userBaseInfo *base.UserBaseInfo) gmevt.UserBaseInfo {
	fashionIDs := []int{}
	for _, id := range userBaseInfo.FashionIDs {
		fashionIDs = append(fashionIDs, int(id))
	}
	return gmevt.UserBaseInfo{
		UID:          int(userBaseInfo.Uid),
		Name:         userBaseInfo.Name,
		Lv:           int(userBaseInfo.Lv),
		UIDIM:        userBaseInfo.UidIM,
		RaceID:       int(userBaseInfo.RaceID),
		ClassID:      int(userBaseInfo.ClassID),
		SpecialistID: int(userBaseInfo.SpecialistID),
		Gender:       int(userBaseInfo.Gender),
		FaceID:       int(userBaseInfo.FaceID),
		HairID:       int(userBaseInfo.HairID),
		HairColorID:  int(userBaseInfo.HairColorID),
		FashionIDs:   fashionIDs,
		MountID:      int(userBaseInfo.MountID),
		BadgeID:      int(userBaseInfo.BadgeID),
		GuildName:    userBaseInfo.GuildName,
	}
}

//////////

//*****************equipconfig begin
//现在不用这种中间件的方法了 直接对接消息
/*
func (DecAction) C2S_Recast(pm proto.Message) *gmevt.ReqRecast {
	m := pm.(*action.C2S_Recast)
	return &gmevt.ReqRecast{
		UID: m.GetUid(),
	}
}
*/
//*****************equipconfig end

func (DecAction) C2S_StudyPaper(pm proto.Message) *gmevt.ReqStudyPaper {
	m := pm.(*action.C2S_StudyPaper)
	// strID := strconv.Itoa(int(m.GetUid()))
	return &gmevt.ReqStudyPaper{
		UID: m.GetUid(),
	}
}

func (DecAction) C2S_Composite(pm proto.Message) *gmevt.ReqComposite {
	m := pm.(*action.C2S_Composite)
	return &gmevt.ReqComposite{
		FormulaID: int(m.GetFormulaId()),
	}
}

func (DecAction) C2S_Identify(pm proto.Message) *gmevt.ReqIdentify {
	m := pm.(*action.C2S_Identify)
	return &gmevt.ReqIdentify{
		UIDs: m.Uids,
	}
}

func (DecAction) C2S_ChangeUserName(pm proto.Message) *gmevt.ReqChangeUserName {
	m := pm.(*action.C2S_ChangeUserName)
	return &gmevt.ReqChangeUserName{
		ItemUID:  m.GetItemUid(),
		UserName: m.GetUserName(),
	}
}

func (DecAction) C2S_UpgradeStove(pm proto.Message) *gmevt.ReqUpgradeStove {
	m := pm.(*action.C2S_UpgradeStove)
	return &gmevt.ReqUpgradeStove{
		NextLevel: int(m.GetNextLevel()),
	}
}

func (DecAction) C2S_UnlockMaskPart(pm proto.Message) *gmevt.ReqUnlockMaskPart {
	m := pm.(*action.C2S_UnlockMaskPart)
	return &gmevt.ReqUnlockMaskPart{
		MaskID: int(m.GetMaskId()),
		Index:  int(m.GetIndex()),
	}
}

func (DecAction) C2S_UnlockMask(pm proto.Message) *gmevt.ReqUnlockMask {
	m := pm.(*action.C2S_UnlockMask)
	return &gmevt.ReqUnlockMask{
		MaskID: int(m.GetMaskId()),
	}
}

func (DecAction) C2S_StartEscort(pm proto.Message) *gmevt.ReqStartEscort {
	m := pm.(*action.C2S_StartEscort)
	return &gmevt.ReqStartEscort{
		NPCUID: m.GetNpcUid(),
		NPCID:  int(m.GetNpcId()),
	}
}

//NOTE: this could return ReqStartFollow or ReqStopFollow
func (DecAction) C2S_Follow(pm proto.Message) event.Event {
	m := pm.(*action.C2S_Follow)
	if m.GetFollowing() {
		return &gmevt.ReqStartFollow{
			NPCUID: m.GetNpcUid(),
		}
	} else {
		return &gmevt.ReqStopFollow{}
	}
}
