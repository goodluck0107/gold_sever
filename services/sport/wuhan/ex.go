package wuhan

func AddUserTicket(hid, uid, actid int64, tcount int64) error {
	sql := `insert into luck_ticket(hid,actid,uid,ticket) values(?,?,?,?) on DUPLICATE KEY UPDATE ticket = ticket + ?`
	err := GetDBMgr().GetDBmControl().Exec(sql, hid, actid, uid, tcount, tcount).Error
	return err
}
