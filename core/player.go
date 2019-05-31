/*
	玩家模块
*/

package core

import (
	"fmt"
	"math/rand"
	"mmogameserver/pb"
	"sync"
	"tinyserver/tsinterface"

	"github.com/golang/protobuf/proto"
)

//Player 玩家类
type Player struct {
	Pid  int32                   //玩家ID
	Conn tsinterface.IConnection //当前玩家的连接
	X    float32                 //平面X轴坐标
	Y    float32                 //高度
	Z    float32                 //平面Y轴坐标
	V    float32                 //玩家的朝向
}

//PidGen playerID生成器
var PidGen int32 = 1 //玩家ID计数

//IDLock 保护PidGen生成器的互斥锁
var IDLock sync.Mutex

//NewPlayer 初始化玩家
func NewPlayer(conn tsinterface.IConnection) *Player {
	//分配一个玩家ID
	IDLock.Lock()
	id := PidGen
	PidGen++
	IDLock.Unlock()

	//创建一个玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), //随机生成玩家上线所在的X轴坐标
		Y:    0,
		Z:    float32(140 + rand.Intn(10)), //随机在140坐标附近Y轴坐标上线
		V:    0,                            //角度为0
	}

	return p
}

//SendMsg 玩家和对端客户端发送消息的方法
func (p *Player) SendMsg(msgID uint32, protoStruct proto.Message) error {
	//将proto结构体数据转换为二进制数据
	banaryProtoData, err := proto.Marshal(protoStruct)
	if err != nil {
		fmt.Println("proto.Marshal error ", err)
		return err
	}

	//调用服务器框架里连接模块给客户端发送数据的方法
	err = p.Conn.Send(msgID, banaryProtoData)
	if err != nil {
		fmt.Println("p.Conn.Send error!", err)
		return err
	}

	return nil
}

//ReturnPid 服务器给客户端发送一个玩家初始ID的方法
func (p *Player) ReturnPid() {
	//定义proto数据：协议里的玩家ID结构体
	protoMsg := &pb.SyncPid{
		Pid: p.Pid,
	}
	//将proto数据作为msgID为1的消息发送给客户端
	p.SendMsg(1, protoMsg)
}

//ReturnPlayerPosition 服务器给客户端发送一个玩家的初始化位置信息的方法
func (p *Player) ReturnPlayerPosition() {
	//构建msgID为200的消息
	protoMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //坐标信息
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将proto数据作为msgID为200的消息发送给客户端
	p.SendMsg(200, protoMsg)
}

//SendTalkMsgToAll 将聊天数据广播给全部在线玩家的方法
func (p *Player) SendTalkMsgToAll(content string) {
	//定义一个符合广播协议数据的protoMsg
	protoMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	//获取全部在线玩家
	players := WorldMgrObj.GetAllPlayers()

	//将protoMsg作为msgID为200的数据广播给全部玩家
	for _, player := range players {
		player.SendMsg(200, protoMsg)
	}
}