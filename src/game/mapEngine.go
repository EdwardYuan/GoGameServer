package game

type MapEngine struct {
	maps map[int]*ObjMap
}

func NewMapEngine() *MapEngine {
	return &MapEngine{
		maps: make(map[int]*ObjMap),
	}
}

func (me *MapEngine) GetMap(mapId int) *ObjMap {
	return me.maps[mapId]
}
