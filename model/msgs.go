package model

import (
	"fmt"
	"github.com/bububa/mymysql/autorc"
	_ "github.com/bububa/mymysql/thrsafe"
	"github.com/nu7hatch/gouuid"
)

type Msg struct {
	ID         string `json:"id"`
	VisitorId  uint64 `json:"visitor_id"`
	KefuId     uint64 `json:"kefu_id"`
	MsgType    uint   `json:"msg_type"`
	Initiator  uint   `json:"initiator"`
	Status     uint   `json:"status"`
	CreateTime string `json:"create_time"`
	SendTime   string `json:"send_time"`
	Content    string `json:"content"`

	DB *autorc.Conn `json:"-"`
}

func (this *Msg) SaveToDb() error {
	if this.DB == nil {
		return NewCscErr("No db connect to save csc_msgs.")
	}
	uuid_val, _ := uuid.NewV4()
	this.ID = uuid_val.String()

	val := fmt.Sprintf("('%s', %d, %d, %d, %d, '%s', '%s', '%s', %d)", this.DB.Escape(this.ID), this.VisitorId, this.KefuId, this.Initiator, this.Status,
		this.DB.Escape(this.CreateTime), this.DB.Escape(this.SendTime), this.DB.Escape(this.Content), this.MsgType)
	_, _, err := this.DB.Query(fmt.Sprintf("INSERT IGNORE INTO csc_msgs (id, visitor_id, kefu_id, initiator, status, create_time, send_time, content, msg_type) VALUES %s", val))

	return err
}
