package static

type MsgAddLeague struct {
	LID      int64 `json:"league_id"`
	Card     int64 `json:"card_num"`
	AreaCode int64 `json:"area_code"`
}

type MsgAddLeageUser struct {
	Uid int64 `json:"uid"`
	LID int64 `json:"league_id"`
}

type MsgFreezeLeague struct {
	LID int64 `json:"league_id"`
}

type MsgUnFreezeLeague struct {
	LID int64 `json:"league_id"`
}

type MsgFreezeLeagueUser struct {
	Uid int64 `json:"uid"`
	LID int64 `json:"league_id"`
}

type MsgUnFreezeLeagueUser struct {
	Uid int64 `json:"uid"`
	LID int64 `json:"league_id"`
}

type LeagueUpdate struct {
	IDs []int64 `json:"ids"`
}

type LeagueUserUpdate struct {
	IDs []int64 `json:"ids"`
}
