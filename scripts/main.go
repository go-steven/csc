package main

import (
	"flag"
	"github.com/bububa/goconfig/config"
	"github.com/bububa/mymysql/autorc"
	_ "github.com/bububa/mymysql/thrsafe"
	log "github.com/kdar/factorlog"
	"os"
	"runtime"
	"time"
)

const (
	_CONFIG_FILE = "/var/code/go/config.cfg"
)

var (
	logFlag = flag.String("log", "", "set log path")

	with_kefu_stats  = flag.Bool("kefu-stats", false, "generate kefu stats")
	with_update_msgs = flag.Bool("update-msgs", true, "update msgs info firstly")
	with_daily_flag  = flag.Bool("daily", true, "as daily")

	logger   *log.FactorLog
	MasterDB *autorc.Conn
)

func init() {
	cfg, _ := config.ReadDefault(_CONFIG_FILE)

	hostMaster, _ := cfg.String("masterdb", "host")
	userMaster, _ := cfg.String("masterdb", "user")
	passwdMaster, _ := cfg.String("masterdb", "passwd")
	dbnameMaster, _ := cfg.String("masterdb", "dbname")

	MasterDB = autorc.New("tcp", "", hostMaster, userMaster, passwdMaster, dbnameMaster)
	MasterDB.Raw.Register("set names utf8")
	MasterDB.Reconnect()
}

func SetGlobalLogger(logPath string) *log.FactorLog {
	sfmt := `%{Color "red:white" "CRITICAL"}%{Color "red" "ERROR"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}:%{ShortFile}:%{Line}] %{Message}%{Color "reset"}`
	logger := log.New(os.Stdout, log.NewStdFormatter(sfmt))
	if len(logPath) > 0 {
		logf, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0640)
		if err != nil {
			return logger
		}
		logger = log.New(logf, log.NewStdFormatter(sfmt))
	}
	logger.SetSeverities(log.INFO | log.WARN | log.ERROR | log.FATAL | log.CRITICAL)
	return logger
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	logger = SetGlobalLogger(*logFlag)

	if *with_kefu_stats {
		if *with_daily_flag { // only generate yesterday's stats
			yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
			if *with_update_msgs {
				if err := update_msgs_stats(MasterDB, yesterday); err != nil {
					logger.Info(err.Error())
					return
				}
			}

			if err := daily_kefu_status(MasterDB, yesterday); err != nil {
				logger.Info(err.Error())
				if err = clear_daily_records(MasterDB, yesterday); err != nil {
					logger.Infof("Rollback failed: %s", err.Error())
				}
			}
		} else { // generate last 30 days stats
			now := time.Now()
			var i time.Duration
			for i = 30; i > 0; i-- {
				stat_date := now.Add((-1 * i * 24) * time.Hour).Format("2006-01-02")
				if *with_update_msgs {
					if err := update_msgs_stats(MasterDB, stat_date); err != nil {
						logger.Info(err.Error())
						break
					}
				}

				if err := daily_kefu_status(MasterDB, stat_date); err != nil {
					logger.Info(err.Error())
					if err = clear_daily_records(MasterDB, stat_date); err != nil {
						logger.Infof("Rollback failed: %s", err.Error())
					}
				}
			}
		}
	}

	return
}
