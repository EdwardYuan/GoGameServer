package game

type MapEngine struct {
	maps map[int]*ObjMap
}

func NewMapEngine() *MapEngine {
	return &MapEngine{
		maps: make(map[int]*ObjMap),
	}
}

func (me *MapEngine) GetMap(mapid int) *ObjMap {
	return me.maps[mapid]
}
