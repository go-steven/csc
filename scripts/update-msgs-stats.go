package main

import (
	"fmt"
	"github.com/bububa/mymysql/autorc"
	"github.com/bububa/mymysql/mysql"
	_ "github.com/bububa/mymysql/thrsafe"
	"time"
)

type UpdateMsgInfo struct {
	MsgId        string
	ResponseTime string
	SessionId    string
}

func update_msgs_stats(DB *autorc.Conn, stat_date string) error {
	query := `SELECT id, visitor_id, kefu_id, initiator, create_time
             FROM csc_msgs
             WHERE create_time >= '%s 00:00:00' AND create_time <= '%s 23:59:59'
             AND msg_type = 1 AND kefu_id > 0 AND visitor_id > 0 
             ORDER BY visitor_id, kefu_id, create_time LIMIT %d OFFSET %d`

	var (
		rows          []mysql.Row
		err           error
		limit, offset int = 500, 0

		lastCreateTime string

		currSessionId             string
		currVisitorId, currKefuId uint64
		currInitiator             uint

		msgId             string
		visitorId, kefuId uint64
		initiator         uint
		createTime        string

		updateMsgs []*UpdateMsgInfo

		duration int64
	)

	for {
		rows, _, err = DB.Query(fmt.Sprintf(query, stat_date, stat_date, limit, offset))
		if err != nil {
			logger.Errorf("Failed to execute SQL: %s, error:%s", query, err.Error())
			return err
		}
		logger.Infof("rows: %d", len(rows))
		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			msgId = row.Str(0)
			visitorId = row.Uint64(1)
			kefuId = row.Uint64(2)
			initiator = row.Uint(3)
			createTime = row.Localtime(4).Format("2006-01-02 15:04:05")

			// 第一条记录
			if currVisitorId == 0 || currKefuId == 0 {
				currVisitorId = visitorId
				currKefuId = kefuId
				currInitiator = initiator

				// 设置当前SESSION，并初始化待保存记录列表
				currSessionId = msgId
				updateMsgs = []*UpdateMsgInfo{}

				// 保存当前记录
				lastCreateTime = createTime
				updateMsgs = append(updateMsgs, &UpdateMsgInfo{
					MsgId:     msgId,
					SessionId: currSessionId,
				})
				continue
			}

			// 会话发生变化
			if visitorId != currVisitorId || kefuId != currKefuId {
				// 将之前的记录保存到数据库
				if err = save_changes_to_db(DB, updateMsgs); err != nil {
					logger.Info(err.Error())
					return err
				}

				currVisitorId = visitorId
				currKefuId = kefuId
				currInitiator = initiator

				// 重置当前SESSION，并初始化待保存记录列表
				currSessionId = msgId
				updateMsgs = []*UpdateMsgInfo{}

				// 保存当前记录
				lastCreateTime = createTime
				updateMsgs = append(updateMsgs, &UpdateMsgInfo{
					MsgId:     msgId,
					SessionId: currSessionId,
				})
				continue
			}
			// 如果距离上一条记录的时间超过预定时间，切换会话
			duration, err = time_between(createTime, lastCreateTime)
			if err != nil {
				return nil
			}
			if duration > 1*30*60 {
				// 将之前的记录保存到数据库
				if err = save_changes_to_db(DB, updateMsgs); err != nil {
					logger.Info(err.Error())
					return err
				}

				// 重置当前SESSION，并初始化待保存记录列表
				currSessionId = msgId
				updateMsgs = []*UpdateMsgInfo{}

				// 保存当前记录
				lastCreateTime = createTime
				updateMsgs = append(updateMsgs, &UpdateMsgInfo{
					MsgId:     msgId,
					SessionId: currSessionId,
				})

				continue
			}

			// 同一会话中，消息发送者发生变化
			if initiator != currInitiator {
				// 更新相关记录：消息响应时间
				update_msgs_info(updateMsgs, createTime)

				currInitiator = initiator
			}

			lastCreateTime = createTime
			updateMsgs = append(updateMsgs, &UpdateMsgInfo{
				MsgId:     msgId,
				SessionId: currSessionId,
			})
		}

		if len(rows) < limit {
			break
		}

		offset = offset + len(rows)
	}

	// 处理未提交的记录
	if currVisitorId != 0 && currKefuId != 0 {
		// 将之前的记录保存到数据库
		if err = save_changes_to_db(DB, updateMsgs); err != nil {
			logger.Info(err.Error())
			return err
		}
	}

	return nil
}

func update_msgs_info(updateMsgs []*UpdateMsgInfo, responseTime string) {
	if responseTime == "" {
		return
	}

	for _, v := range updateMsgs {
		if v.ResponseTime == "" {
			v.ResponseTime = responseTime
		}
	}
}

func save_changes_to_db(DB *autorc.Conn, updateMsgs []*UpdateMsgInfo) error {
	query := `UPDATE csc_msgs SET response_time = '%s', session_id = '%s' WHERE id = '%s'`
	//logger.Infof("Execute UPDATE sql: %s", query)
	//logger.Infof("data: %v", updateMsgs)

	var (
		responseTime, sessionId string
	)

	cnt := 0
	for _, v := range updateMsgs {

		responseTime = v.ResponseTime
		sessionId = v.SessionId
		if responseTime == "" && sessionId == "" {
			continue
		}

		_, result, err := DB.Query(fmt.Sprintf(query, responseTime, sessionId, v.MsgId))
		if err != nil || result == nil {
			logger.Errorf("Failed to execute SQL: %s, error:%s", query, err.Error())
			return err
		}

		// after update 200 rows in database, wait 1 seconds
		cnt = cnt + 1
		if cnt >= 200 {
			time.Sleep(1 * time.Second)
			cnt = 0
		}
	}

	return nil
}

func time_between(start, end string) (int64, error) {
	var (
		start_time, end_time time.Time
		err                  error
	)

	start_time, err = time.Parse("2006-01-02 15:04:05", start)
	if err != nil {
		return 0, err
	}

	end_time, err = time.Parse("2006-01-02 15:04:05", end)
	if err != nil {
		return 0, err
	}

	return end_time.Unix() - start_time.Unix(), nil
}
