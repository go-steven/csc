package model

import (
	"errors"
	"fmt"
	"github.com/bububa/mymysql/autorc"
	_ "github.com/bububa/mymysql/thrsafe"
	"sync"
	"time"
)

const (
	UID = iota
	NAME
	NICK
	AFFILIATE
	ROLE
	MANAGER
	SIGNIN_AT
	SIGNUP_AT
	STATUS
	SESSION
	KEFU_ID
	KEFU_NAME
	REAL_KEFU_ID
	REAL_KEFU_NAME
	CHAT_STATUS
	CHAT_CNT
	SC_LEVEL
	DEPOSIT_CNT
)

const (
	ONLINE, OFFLINE uint = 0, 1
	VISITOR, KEFU   uint = 1, 2
)

type User struct {
	uid          uint64
	name         string
	nick         string
	affiliate    string
	role         uint
	manager      uint
	signinAt     string
	signupAt     string
	status       uint
	session      string
	kefuId       uint64
	kefuName     string
	realKefuId   uint64
	realKefuName string
	chatStatus   uint
	chatCnt      uint
	scLevel      uint
	depositCnt   uint

	db *autorc.Conn
	m  *sync.RWMutex
}

type UserCopy struct {
	Uid          uint64 `json:"uid"`
	Name         string `json:"name"`
	Nick         string `json:"nick"`
	Affiliate    string `json:"affiliate"`
	Role         uint   `json:"role"`
	Manager      uint   `json:"manager"`
	SigninAt     string `json:"signin_at"`
	SignupAt     string `json:"signup_at"`
	Status       uint   `json:"status"`
	Session      string `json:"session"`
	KefuId       uint64 `json:"last_kefu_id"`
	KefuName     string `json:"last_kefu_name"`
	RealKefuId   uint64 `json:"curr_kefu_id"`
	RealKefuName string `json:"curr_kefu_name"`
	ChatStatus   uint   `json:"curr_chat_sts"`
	ChatCnt      uint   `json:"chat_cnt"`
	ScLevel      uint   `json:"sc_level"`
	DepositCnt   uint   `json:"deposit_cnt"`
}

func NewUser(uid uint64, role uint, db *autorc.Conn) *User {
	userExtInfo, err := GetUserExtInfo(uid, role, db)
	if err != nil || userExtInfo.Name == "" {
		return nil
	}

	if userExtInfo.Nick == "" {
		if role == KEFU {
			userExtInfo.Nick = fmt.Sprintf("客服%d", uid)
		} else {
			userExtInfo.Nick = userExtInfo.Name
		}
	}
	return &User{
		m:          new(sync.RWMutex),
		uid:        uid,
		name:       userExtInfo.Name,
		nick:       userExtInfo.Nick,
		affiliate:  GetAffilateHead(userExtInfo.Affiliate),
		kefuId:     userExtInfo.KefuId,
		role:       role,
		manager:    userExtInfo.Manager,
		status:     userExtInfo.Status,
		db:         db,
		chatStatus: 1,
		scLevel:    userExtInfo.ScLevel,
		depositCnt: userExtInfo.DepositCnt,
	}
}

func NewUserModel() *User {
	return &User{
		m: new(sync.RWMutex),
	}
}

func (this *User) Copy() *UserCopy {
	this.m.RLock()
	defer this.m.RUnlock()

	return &UserCopy{
		Uid:          this.uid,
		Name:         this.name,
		Nick:         this.nick,
		Affiliate:    this.affiliate,
		Role:         this.role,
		Manager:      this.manager,
		SigninAt:     this.signinAt,
		SignupAt:     this.signinAt,
		Status:       this.status,
		Session:      this.session,
		KefuId:       this.kefuId,
		KefuName:     this.kefuName,
		RealKefuId:   this.realKefuId,
		RealKefuName: this.realKefuName,
		ChatStatus:   this.chatStatus,
		ChatCnt:      this.chatCnt,
		ScLevel:      this.scLevel,
		DepositCnt:   this.depositCnt,
	}
}

func (this *User) GetStrField(field int) string {
	this.m.RLock()
	defer this.m.RUnlock()

	var result string
	switch field {
	case NAME:
		result = this.name
	case NICK:
		result = this.nick
	case AFFILIATE:
		result = this.affiliate
	case SESSION:
		result = this.session
	case KEFU_NAME:
		result = this.kefuName
	case REAL_KEFU_NAME:
		result = this.realKefuName
	case SIGNIN_AT:
		result = this.signinAt
	case SIGNUP_AT:
		result = this.signupAt
	}

	return result
}

func (this *User) SetStrField(field int, val string) error {
	this.m.Lock()
	defer this.m.Unlock()

	switch field {
	case NAME:
		this.name = val
	case NICK:
		this.nick = val
	case AFFILIATE:
		this.affiliate = val
	case SESSION:
		this.session = val
	case KEFU_NAME:
		this.kefuName = val
	case REAL_KEFU_NAME:
		this.realKefuName = val
	case SIGNIN_AT:
		this.signinAt = val
	case SIGNUP_AT:
		this.signupAt = val
	default:
		return errors.New("unknow string field: " + string(field))
	}

	return nil
}

func (this *User) GetUintField(field int) uint {
	this.m.RLock()
	defer this.m.RUnlock()

	var result uint
	switch field {
	case ROLE:
		result = this.role
	case MANAGER:
		result = this.manager
	case STATUS:
		result = this.status
	case CHAT_STATUS:
		result = this.chatStatus
	case CHAT_CNT:
		result = this.chatCnt
	case SC_LEVEL:
		result = this.scLevel
	case DEPOSIT_CNT:
		result = this.depositCnt
	}

	return result
}

func (this *User) SetUintField(field int, val uint) error {
	this.m.Lock()
	defer this.m.Unlock()

	switch field {
	case ROLE:
		this.role = val
	case MANAGER:
		this.manager = val
	case STATUS:
		this.status = val
	case CHAT_STATUS:
		this.chatStatus = val
	case CHAT_CNT:
		this.chatCnt = val
	case SC_LEVEL:
		this.scLevel = val
	case DEPOSIT_CNT:
		this.depositCnt = val
	default:
		return errors.New("unknow uint field: " + string(field))
	}

	return nil
}

func (this *User) GetUint64Field(field int) uint64 {
	this.m.RLock()
	defer this.m.RUnlock()

	var result uint64
	switch field {
	case UID:
		result = this.uid
	case KEFU_ID:
		result = this.kefuId
	case REAL_KEFU_ID:
		result = this.realKefuId
	}

	return result
}

func (this *User) SetUint64Field(field int, val uint64) error {
	this.m.Lock()
	defer this.m.Unlock()

	switch field {
	case UID:
		this.uid = val
	case KEFU_ID:
		this.kefuId = val
	case REAL_KEFU_ID:
		this.realKefuId = val
	default:
		return errors.New("unknow uint field: " + string(field))
	}

	return nil
}

func (this *User) Offline() {
	this.m.Lock()
	defer this.m.Unlock()

	this.status = OFFLINE
	this.signupAt = time.Now().Format("2006-01-02 15:04:05")
}

func (this *User) Online(session string) {
	this.m.Lock()
	defer this.m.Unlock()

	this.session = session
	this.status = ONLINE
	this.signinAt = time.Now().Format("2006-01-02 15:04:05")
	this.signupAt = ""
}

func (this *User) SetKefu(kefu *User) {
	this.m.Lock()
	defer this.m.Unlock()

	this.kefuId = kefu.GetUint64Field(UID)
	this.kefuName = kefu.GetStrField(NAME)
	this.realKefuId = this.kefuId
	this.realKefuName = this.kefuName
}

func (this *User) SaveToDb() error {
	//logger.Infof("Save User to DB.")
	if this.db == nil {
		return NewCscErr("No db connect to save csc_users.")
	}

	this.m.RLock()
	defer this.m.RUnlock()

	//logger.Infof("uid: %d, manager:%d", this.uid, this.manager)

	val := fmt.Sprintf("(%d, '%s', '%s', '%s', %d, %d, '%s', '%s', %d, '%s', %d, %d, %d)", this.uid, this.db.Escape(this.name), this.db.Escape(this.nick), this.db.Escape(this.affiliate), this.role, this.manager,
		this.db.Escape(this.signinAt), this.db.Escape(this.signupAt), this.status, this.db.Escape(this.session), this.kefuId, this.scLevel, this.depositCnt)
	_, _, err := this.db.Query(fmt.Sprintf("INSERT IGNORE INTO csc_users (uid, name, nick, affiliate, role, manager, signin_at, signup_at, status, session, kefu_id, sc_level, deposit_cnt) VALUES %s ON DUPLICATE KEY UPDATE name=VALUES(name), nick=VALUES(nick), affiliate=VALUES(affiliate), status=VALUES(status), session=VALUES(session), signin_at=VALUES(signin_at), signup_at=VALUES(signup_at), kefu_id=VALUES(kefu_id), manager=VALUES(manager), sc_level=VALUES(sc_level), deposit_cnt=VALUES(deposit_cnt)", val))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (this *User) UpdateInfoFromDb() error {
	query := `SELECT u.name, u.nick, u.affiliate, u.kefu_id, k.name as kefu_name, u.status, u.manager, u.sc_level, u.deposit_cnt FROM csc_users u 
			 LEFT JOIN csc_users k ON (k.uid = u.kefu_id AND k.role = 2 AND u.role = 1) 
			 WHERE u.uid = %d AND u.role = %d LIMIT 1`
	row, _, err := this.db.QueryFirst(query, this.uid, this.role)
	if err != nil {
		logger.Error(err)
		return err
	}

	if row != nil {
		this.m.Lock()
		defer this.m.Unlock()

		this.name = row.Str(0)
		this.nick = row.Str(1)
		this.affiliate = row.Str(2)
		this.kefuId = row.Uint64(3)
		this.kefuName = row.Str(4)
		this.status = row.Uint(5)
		this.manager = row.Uint(6)
		this.scLevel = row.Uint(7)
		this.depositCnt = row.Uint(8)
	}

	return nil
}

func (this *User) UpdateExtInfo() error {
	userExtInfo, err := GetUserExtInfo(this.uid, this.role, this.db)
	if err != nil {
		return err
	}

	this.m.Lock()
	this.scLevel = userExtInfo.ScLevel
	this.depositCnt = userExtInfo.DepositCnt
	this.m.Unlock()

	if err = this.SaveToDb(); err != nil {
		return err
	}

	return nil
}

func GetUserExtInfo(uid uint64, role uint, db *autorc.Conn) (userExtInfo *UserCopy, myErr error) {
	userExtInfo = &UserCopy{}

	var query string
	if role == VISITOR {
		query = `SELECT
									        IFNULL(u.wangwang, IFNULL(a.nick, za.nick)) AS nick,
									        c.nick as csc_nick,
								            zf.title AS affiliate,
								            c.kefu_id,
								            c.status,
								            c.manager,
								            a.sc_level,
								            (
								                SELECT
								                    IFNULL(count(zt.trade_id), 0)
								                FROM
								                    zhanwai_trades zt
								                WHERE
								                    zt.user_id = u.id
								                AND
								                    zt.status = 1
								            ) AS deposit_cnt   								            
										FROM
										  users u
										LEFT JOIN csc_users c ON c.uid = u.id AND c.role = %d
										LEFT JOIN zhanwai_users zu ON zu.user_id = u.id
										LEFT JOIN zhanwai_affiliate AS zf ON zf.id = zu.affiliate
										LEFT JOIN zhanwai_accounts AS za ON za.user_id = zu.user_id
										LEFT JOIN accounts AS a ON a.id = za.account_id
										WHERE u.id = %d LIMIT 1`
	} else {
		query = `SELECT
									        IFNULL(u.realname, u.name) AS nick,
									        c.nick as csc_nick,
								            '喜宝DSP' AS affiliate,
								            0 as kefu_id,
								            c.status,
								            c.manager,
								            c.sc_level,
								            c.deposit_cnt
										FROM
										  users u
										LEFT JOIN csc_users c ON c.uid = u.id AND c.role=%d
										WHERE u.id = %d LIMIT 1`
	}
	row, _, err := db.QueryFirst(query, role, uid)
	if err != nil {
		logger.Error(err)
		myErr = err
		return
	}

	if row != nil {
		userExtInfo.Name = row.Str(0)
		userExtInfo.Nick = row.Str(1)
		userExtInfo.Affiliate = row.Str(2)
		userExtInfo.KefuId = row.Uint64(3)
		userExtInfo.Status = row.Uint(4)
		userExtInfo.Manager = row.Uint(5)
		userExtInfo.ScLevel = row.Uint(6)
		userExtInfo.DepositCnt = row.Uint(7)
	}

	return
}

func GetAffilateHead(affiliate string) string {
	if affiliate == "" {
		return ""
	}

	if affiliate == "京东DSP" || affiliate == "JD" {
		return "JD"
	}

	r := []rune(affiliate)
	return string(r[0])
}
