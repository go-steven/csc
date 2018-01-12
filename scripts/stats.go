package main

import (
	"errors"
	"fmt"
	"github.com/bububa/mymysql/autorc"
	"github.com/bububa/mymysql/mysql"
	_ "github.com/bububa/mymysql/thrsafe"
	"strings"
)

var daily_stats_sqls = map[string]string{
	// 清除已有记录
	"clear_records": `DELETE FROM csc_kefu_stats
        WHERE stat_date = '%s'`,

	// 查询： 客服， 日期，接待总数
	"get_total_visitors": `SELECT 
             kefu_id, 
             COUNT(DISTINCT visitor_id) AS total_visitors
         FROM csc_msgs
           WHERE create_time >= '%s 00:00:00' AND create_time <= '%s 23:59:59'
             AND msg_type = 1 
             AND kefu_id > 0 and visitor_id > 0
         GROUP BY kefu_id
         ORDER BY kefu_id`,

	// 查询： 客服  日期  重复接待数
	"get_dup_visitors": `SELECT 
                        A.kefu_id, 
                        COUNT(DISTINCT A.visitor_id) AS dup_visitors
                    FROM csc_msgs A
                    WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                       AND A.msg_type = 1 
                       AND A.kefu_id > 0 AND A.visitor_id > 0
                       AND EXISTS (
                            SELECT 1 FROM csc_msgs B 
                            WHERE B.create_time >= '%s 00:00:00' AND B.create_time <= '%s 23:59:59'
                                AND B.msg_type = 1 
                                AND B.kefu_id > 0 AND B.visitor_id > 0
                                AND B.visitor_id = A.visitor_id 
                                AND B.kefu_id <> A.kefu_id 
                                AND (B.kefu_id <> 61476 OR A.kefu_id = 61476)
                        )   
                    GROUP BY A.kefu_id
                    ORDER BY A.kefu_id`,

	// 查询： 客服  日期  未回复客户数
	"get_no_response_visitors": `SELECT 
                        A.kefu_id, 
                        COUNT(DISTINCT A.visitor_id) AS no_response_visitors
                    FROM csc_msgs A
                    WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                       AND A.msg_type = 1 
                       AND A.initiator = 1
                       AND A.kefu_id > 0 AND A.visitor_id > 0
                       AND A.response_time = '0000-00-00 00:00:00'
                    GROUP BY A.kefu_id
                    ORDER BY A.kefu_id`,

	// 查询：客服  日期  客户结束的对话数量
	"get_visitor_end": `SELECT 
                    M.kefu_id, 
                    COUNT(distinct M.visitor_id) AS visitor_end
                FROM csc_msgs M 
                WHERE M.create_time >= '%s 00:00:00' AND M.create_time <= '%s 23:59:59'
                    AND M.msg_type = 1 
                    AND M.kefu_id > 0 AND M.visitor_id > 0
                    AND M.initiator = 1 
                    AND EXISTS (
                        SELECT 1 FROM ( 
                            SELECT A.visitor_id, A.kefu_id, MAX(A.create_time) AS create_time
                            FROM csc_msgs A
                            WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                                AND A.msg_type = 1 
                                AND A.kefu_id > 0 AND A.visitor_id > 0
                            GROUP BY A.visitor_id, A.kefu_id
                        ) B 
                        WHERE B.visitor_id = M.visitor_id 
                            AND B.kefu_id = M.kefu_id 
                            AND B.create_time = M.create_time
                    )
                    GROUP BY M.kefu_id
                    ORDER BY M.kefu_id`,

	// 查询：客服  日期  总接待时长
	"get_avg_duration": `SELECT 
                                B.kefu_id, 
                                ROUND(AVG(B.duration)) AS avg_duration 
                    FROM ( 
                        SELECT  
                            A.kefu_id, 
                            A.session_id,
                            UNIX_TIMESTAMP(MAX(A.response_time)) - UNIX_TIMESTAMP(MIN(A.response_time)) as duration
                        FROM csc_msgs A
                        WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                            AND A.msg_type = 1
                            AND A.initiator = 1 
                            AND A.kefu_id > 0 AND A.visitor_id > 0
                            AND A.session_id is not null AND A.response_time >= '%s 00:00:00'
                        GROUP BY A.kefu_id, A.session_id
                    ) B
                    GROUP BY B.kefu_id
                    ORDER BY B.kefu_id`,

	// 查询：客服  日期  平均回复时长
	"get_avg_response_duration": `SELECT 
                                B.kefu_id, 
                                ROUND(AVG(B.duration)) AS avg_duration 
                    FROM ( 
                        SELECT  
                            A.kefu_id, 
                            A.session_id,
                            UNIX_TIMESTAMP(MIN(A.response_time)) - UNIX_TIMESTAMP(MIN(A.create_time)) as duration
                        FROM csc_msgs A
                        WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                            AND A.msg_type = 1
                            AND A.initiator = 1 
                            AND A.kefu_id > 0 AND A.visitor_id > 0
                            AND A.session_id is not null AND A.response_time >= '%s 00:00:00'
                        GROUP BY A.kefu_id, A.session_id
                    ) B
                    GROUP BY B.kefu_id
                    ORDER BY B.kefu_id`,

	// 查询：客服  日期  客户总咨询字数
	"get_total_visitor_content": `SELECT 
                        B.kefu_id, 
                        SUM(B.content_len) AS total_visitor_content 
                    FROM (
                        SELECT 
                            A.visitor_id, 
                            A.kefu_id, 
                            SUM(CHAR_LENGTH(A.content)) AS content_len 
                        FROM csc_msgs A
                        WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                            AND A.msg_type = 1 
                            AND A.kefu_id > 0 AND A.visitor_id > 0
                            AND A.initiator = 1
                        GROUP BY A.visitor_id, A.kefu_id
                    ) B
                    GROUP BY B.kefu_id
                    ORDER BY B.kefu_id`,

	// 查询：客服  日期  客服总回复字数
	"get_total_kefu_content": `SELECT 
                        B.kefu_id, 
                        SUM(B.content_len) AS total_kefu_content 
                    FROM (
                        SELECT 
                            A.visitor_id, 
                            A.kefu_id, 
                            SUM(CHAR_LENGTH(A.content)) AS content_len 
                        FROM csc_msgs A
                        WHERE A.create_time >= '%s 00:00:00' AND A.create_time <= '%s 23:59:59'
                            AND A.msg_type = 1 
                            AND A.kefu_id > 0 AND A.visitor_id > 0
                            AND A.initiator = 2
                        GROUP BY A.visitor_id, A.kefu_id
                    ) B
                    GROUP BY B.kefu_id
                    ORDER BY B.kefu_id`,
}

type KefuStats struct {
	KefuId              uint64 `json:"kefu_id"`
	StatDate            string `json:"stat_date"`
	TotalVisitors       uint64 `json:"total_visitors"`
	DupVisitors         uint64 `json:"dup_visitors"`
	NoResponseVisitors  uint64 `json:"no_response_visitors"`
	VisitorEnd          uint64 `json:"visitor_end"`
	AvgDuration         uint64 `json:"avg_duration"`
	AvgResponseDuration uint64 `json:"avg_response_duration"`
	AvgVisitorContent   uint64 `json:"avg_visitor_content"`
	AvgKefuContent      uint64 `json:"avg_kefu_content"`
}

func daily_kefu_status(DB *autorc.Conn, stat_date string) error {
	logger.Infof("Start to generate daily kefu stats: %s", stat_date)
	defer logger.Infof("Generate daily kefu stats ended: %s", stat_date)

	var (
		err                              error
		totalVisitors, dupVisitors       map[uint64]uint64
		avgResponseDuration, avgDuration map[uint64]uint64
		noResponseVisitors, visitorEnd   map[uint64]uint64
		visitorContent, kefuContent      map[uint64]uint64
	)

	// 清除已有记录
	if err = clear_daily_records(DB, stat_date); err != nil {
		return nil
	}

	totalVisitors, err = get_stats(DB, "get_total_visitors", stat_date)
	if err != nil {
		return err
	}

	dupVisitors, err = get_stats(DB, "get_dup_visitors", stat_date)
	if err != nil {
		return err
	}

	noResponseVisitors, err = get_stats(DB, "get_no_response_visitors", stat_date)
	if err != nil {
		return err
	}

	visitorEnd, err = get_stats(DB, "get_visitor_end", stat_date)
	if err != nil {
		return err
	}

	avgDuration, err = get_stats(DB, "get_avg_duration", stat_date)
	if err != nil {
		return err
	}

	avgResponseDuration, err = get_stats(DB, "get_avg_response_duration", stat_date)
	if err != nil {
		return err
	}

	visitorContent, err = get_stats(DB, "get_total_visitor_content", stat_date)
	if err != nil {
		return err
	}

	kefuContent, err = get_stats(DB, "get_total_kefu_content", stat_date)
	if err != nil {
		return err
	}

	var (
		stat      uint64
		ok        bool
		kefuStats *KefuStats
	)

	resultList := []*KefuStats{}

	for k, v := range totalVisitors {

		kefuStats = &KefuStats{
			KefuId:        k,
			StatDate:      stat_date,
			TotalVisitors: v,
		}

		if stat, ok = dupVisitors[k]; ok {
			kefuStats.DupVisitors = stat
		}

		if stat, ok = noResponseVisitors[k]; ok {
			kefuStats.NoResponseVisitors = stat
		}

		if stat, ok = avgDuration[k]; v > 0 && ok {
			kefuStats.AvgDuration = stat
		}

		if stat, ok = avgResponseDuration[k]; v > 0 && ok {
			kefuStats.AvgResponseDuration = stat
		}

		if stat, ok = visitorEnd[k]; ok {
			kefuStats.VisitorEnd = stat
		}

		if stat, ok = visitorContent[k]; v > 0 && ok {
			kefuStats.AvgVisitorContent = division_round(stat, v)
		}

		if stat, ok = kefuContent[k]; v > 0 && ok {
			kefuStats.AvgKefuContent = division_round(stat, v)
		}

		resultList = append(resultList, kefuStats)
	}

	if err = saveKefuStats(DB, resultList); err != nil {
		return err
	}

	return nil
}

func saveKefuStats(DB *autorc.Conn, data []*KefuStats) error {
	if len(data) == 0 {
		return nil
	}

	queryInsert := `INSERT INTO csc_kefu_stats (kefu_id, stat_date, total_visitors, dup_visitors, no_response_visitors, visitor_end, avg_duration, avg_response_duration, avg_visitor_content, avg_kefu_content) VALUES %s`

	valInsert := []string{}

	for _, v := range data {
		valInsert = append(valInsert, fmt.Sprintf("(%d, '%s', %d, %d, %d, %d, %d, %d, %d, %d)",
			v.KefuId, v.StatDate, v.TotalVisitors, v.DupVisitors, v.NoResponseVisitors, v.VisitorEnd,
			v.AvgDuration, v.AvgResponseDuration, v.AvgVisitorContent, v.AvgKefuContent))
	}

	if len(valInsert) > 0 {
		_, _, err := DB.Query(queryInsert, strings.Join(valInsert, ","))
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}
func clear_daily_records(DB *autorc.Conn, stat_date string) error {
	_, result, err := DB.Query(daily_stats_sqls["clear_records"], stat_date)
	if err != nil || result == nil {
		logger.Errorf("Failed to execute SQL: %s, error:%s", daily_stats_sqls["clear_records"], err.Error())
		return err
	}

	return nil
}

func get_stats(DB *autorc.Conn, sql_id, stat_date string) (map[uint64]uint64, error) {
	if _, ok := daily_stats_sqls[sql_id]; !ok {
		return nil, errors.New(fmt.Sprintf("No SQL found for %s", sql_id))
	}

	query := daily_stats_sqls[sql_id]

	var (
		rows []mysql.Row
		err  error
	)
	switch sql_id {
	case "get_total_visitors", "get_total_visitor_content", "get_total_kefu_content", "get_no_response_visitors":
		rows, _, err = DB.Query(fmt.Sprintf(query, stat_date, stat_date))
	case "get_avg_duration", "get_avg_response_duration":
		rows, _, err = DB.Query(fmt.Sprintf(query, stat_date, stat_date, stat_date))
	case "get_dup_visitors", "get_visitor_end":
		rows, _, err = DB.Query(fmt.Sprintf(query, stat_date, stat_date, stat_date, stat_date))
	}
	if err != nil {
		logger.Errorf("Failed to execute SQL: %s, error:%s", query, err.Error())
		return nil, err
	}

	retMap := make(map[uint64]uint64)
	for _, row := range rows {
		retMap[row.Uint64(0)] = row.Uint64(1)
	}

	return retMap, nil
}

func division_round(divisor, dividend uint64) uint64 {
	result := divisor / dividend
	remainder := divisor % dividend
	mid := dividend / 2

	if remainder > mid || (dividend%2 == 0 && remainder == mid) {
		return result + 1
	}

	return result
}
