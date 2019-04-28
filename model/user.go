package model

import (
	"fmt"
	"os"
	"time"

	"UULoServer/lib"
	"UULoServer/logs"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

//游戏内用户表
type GameUser struct {
	ObjectId string `xorm:"pk 'objectid'"` //主键，该条记录的唯一标识
	Nickname string `xorm:"nickname"`      //用户名称
	Sex      int    `xorm:"sex"`           //用户性别，1为男性，2为女性
	Province string `xorm:"province"`      //用户个人资料填写的省份
	City     string `xorm:"city"`          //用户个人资料填写的城市
	Country  string `xorm:"country"`       //国家，如中国为CN

	//---------------------------------------------------
	Tp    string `xorm:"type"`  //账号类型
	Ud    int    `xorm:"ud"`    //钻石 或对应 虚拟币
	Cw    int    `xorm:"cw"`    //连胜
	Br    int    `xorm:"br"`    //破产数
	Guide int    `xorm:"guide"` //引导进度
	Scene int    `xorm:"scene"` //选择的场景
	//---------------------------------------------------

	Headimgurl  string    `xorm:"headimgurl"` //用户头像的URL地址
	Unionid     string    `xorm:"unionid"`    //用户统一标识
	Img         int       `xorm:"img"`        //形象ID
	Up          int       `xorm:"up"`         //金币
	Ct          int       `xorm:"ct"`         //对局总数
	Wn          int       `xorm:"wn"`         //胜利次数
	Lastsubsidy time.Time `xorm:lastsubsidy`  // 上次补助时间
	CreatedAt   time.Time `xorm:"created 'createdAt'"`
	UpdatedAt   time.Time `xorm:"updated 'updatedAt'"`
}

//应用内用户表
type User struct {
	Id       string `xorm:"pk 'id'"`  //主键	 对应游戏表  unionid
	Username string `xorm:"username"` //用户名 对应游戏表  nickname
}

//房间配置表
type RoomCfg struct {
	Id            int `xorm:"pk 'id'"`       //房间id
	Underpoint    int `xorm:"underpoint"`    //底分
	Initmul       int `xorm:"initmul"`       //初始倍数
	Minenterpoint int `xorm:"minenterpoint"` //最小进入限制
	Maxenterpoint int `xorm:"maxenterpoint"` //最大进入限制
	Ticket        int `xorm:"ticket"`        //门票
	Maxearn       int `xorm:"maxearn"`       //收益封顶
}

//排行结果
////名字 |  胜局 wn | 连胜 cw |  3   |  4  | 头像 img
type Ranking struct {
	Username string
	Wn       int
	Cw       int
	Up       int
	Ct       int
	Img      int
}

var engine *xorm.Engine
var roomscfg []RoomCfg

func init() {
	logs.SetLogFile("uulandlord")
	var err error
	// engine, err = xorm.NewEngine("mysql",
	// 	"developer:dev783edg@tcp(rm-uf65643l032z4k37qo.mysql.rds.aliyuncs.com:3306)/uulandloard")
	engine, err = xorm.NewEngine("mysql", "root:123456@tcp(127.0.0.1:3306)/uulandloard")
	err = engine.Ping()
	if err != nil {
		logs.Error("open db error:%s", err.Error())
		os.Exit(-1)
	}
	logs.Info("open db success.")
}

//orm指定表名
func (this *GameUser) TableName() string {
	return "game_ddz"
}

func (this *User) TableName() string {
	return "user"
}

func (this *RoomCfg) TableName() string {
	return "goldroom"
}

func InitRoomsCfg() []RoomCfg {
	roomscfg = make([]RoomCfg, 0)
	err := engine.Find(&roomscfg)

	if err != nil {
		logs.Error("open room cfg error:%s", err.Error())
		return nil
	} else {
		return roomscfg
	}
}

func GetDetialCfg(id int) *RoomCfg {

	for _, k := range roomscfg {
		if k.Id == id {
			return &k
		}
	}

	return nil
}

//取得房间配置信息
func GetRoomsCfg() []string {

	str := make([]string, len(roomscfg))
	// []string{"1|100|1|10000|100000|1|1000000", "2|1000|5|10000|10000000|1|10000000"}
	// id            int `xorm:"pk 'id'"`       //房间id
	// underpoint    int `xorm:"underpoint"`    //底分
	// initmul       int `xorm:"initmul"`       //初始倍数
	// minenterpoint int `xorm:"minenterpoint"` //最小进入限制
	// maxenterpoint int `xorm:"maxenterpoint"` //最大进入限制
	// ticket        int `xorm:"ticket"`        //门票
	// maxearn       int `xorm:"maxearn"`       //收益封顶
	for i, k := range roomscfg {
		s := fmt.Sprintf("%d|%d|%d|%d|%d|%d|%d", k.Id, k.Underpoint, k.Initmul, k.Minenterpoint, k.Maxenterpoint, k.Ticket, k.Maxearn)
		//str = append(str, s)
		str[i] = s
	}
	return str
}

//创建一个用户
func NewUser(name, province, city, country, headimg, uid string, sex int) (*GameUser, error) {
	obj := lib.NewObjectId()
	u := &GameUser{ObjectId: obj, Nickname: name, Sex: sex, Province: province,
		City: city, Country: country, Headimgurl: headimg, Unionid: uid,
		Img: 1001, Up: 10000, Ct: 0, Wn: 0,
		Tp: "1", Ud: 0, Cw: 0, Br: 0, Guide: 0, Scene: 3001, Lastsubsidy: time.Unix(0, 0)}

	/*
			Type  string `xorm:"type"`  //账号类型
		Ud    int    `xorm:"ud"`    //钻石 或对应 虚拟币
		Cw    int    `xorm:"cw"`    //连胜
		Br    int    `xorm:"br"`    //破产数
		Guide int    `xorm:"guide"` //引导进度
		Scene int    `xorm:"scene"` //选择的场景
	*/

	_, err := engine.InsertOne(u)
	return u, err
}

//根据Unionid  从游戏用户表中获取用户
//return *User 用户  bool 是否存在
func GetUser(Unionid string) (*GameUser, bool) {
	u := &GameUser{}
	has, _ := engine.Where("unionid=?", Unionid).Get(u)
	return u, has
}

//从基础用户表中查找用户
func GetBaseUser(id string) (*User, bool) {
	u := &User{}
	has, _ := engine.Where("id=?", id).Get(u)
	return u, has
}

//更新用户
func UpdateUser(objid string, bm map[string]interface{}) error {
	_, err := engine.Table(new(GameUser)).Id(objid).Update(bm)

	return err
}

//以金币排名,取前3名
func GetRank(ranktype int) ([]string, error) {
	//名字 |  胜局 wn | 连胜 cw |  3   |  4  | 头像 img
	us := make([]GameUser, 0)

	var err error
	if ranktype == 1 {
		err = engine.Cols("nickname", "wn", "cw", "up", "ct", "img").Desc("up").Limit(3, 0).Find(&us)
	} else {
		err = engine.Cols("nickname", "wn", "cw", "up", "ct", "img").Desc("ct").Limit(3, 0).Find(&us)
	}

	rank := make([]string, len(us))
	for i, v := range us {
		rank[i] = fmt.Sprintf("%s|%d|%d|%d|%d|%d", v.Nickname, v.Wn, v.Cw, v.Up, v.Ct, v.Img)
	}
	return rank, err
}
