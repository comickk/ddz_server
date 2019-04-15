package model

import (
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

	Headimgurl string    `xorm:"headimgurl"` //用户头像的URL地址
	Unionid    string    `xorm:"unionid"`    //用户统一标识
	Img        int       `xorm:"img"`        //形象ID
	Up         int       `xorm:"up"`         //金币
	Ct         int       `xorm:"ct"`         //对局总数
	Wn         int       `xorm:"wn"`         //胜利次数
	CreatedAt  time.Time `xorm:"created 'createdAt'"`
	UpdatedAt  time.Time `xorm:"updated 'updatedAt'"`
}

//应用内用户表
type User struct {
	Id       string `xorm:"pk 'id'"`  //主键	 对应游戏表  unionid
	Username string `xorm:"username"` //用户名 对应游戏表  nickname
}

//排行结果
type Ranking struct {
	Username string
	Up       int
}

var engine *xorm.Engine

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

//创建一个用户
func NewUser(name, province, city, country, headimg, uid string, sex int) (*GameUser, error) {
	obj := lib.NewObjectId()
	u := &GameUser{ObjectId: obj, Nickname: name, Sex: sex, Province: province,
		City: city, Country: country, Headimgurl: headimg, Unionid: uid,
		Img: 1001, Up: 10000, Ct: 0, Wn: 0,
		Tp: "1", Ud: 0, Cw: 0, Br: 0, Guide: 0, Scene: 3001}

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

//以金币排名,取前20名
func RankingByUp() ([]Ranking, error) {
	us := make([]GameUser, 0)
	err := engine.Cols("nickname", "up").Desc("up").Limit(20, 0).Find(&us)
	rs := make([]Ranking, len(us))
	if err != nil {
		return rs, err
	}
	for i, v := range us {
		rs[i].Username = v.Nickname
		rs[i].Up = v.Up
	}
	return rs, err
}
