package protocol

type Protocol int

// 发给worker，gateway有一个新的连接
const CMD_ON_CONNECT = Protocol(1)

// 发给worker的，客户端有消息
const CMD_ON_MESSAGE = Protocol(3)

// 发给worker上的关闭链接事件
const CMD_ON_CLOSE = Protocol(4)

// 发给gateway的向单个用户发送数据
const CMD_SEND_TO_ONE = Protocol(5)

// 发给gateway的向所有用户发送数据
const CMD_SEND_TO_ALL = Protocol(6)

// 发给gateway的踢出用户
// 1、如果有待发消息，将在发送完后立即销毁用户连接
// 2、如果无待发消息，将立即销毁用户连接
const CMD_KICK = Protocol(7)

// 发给gateway的立即销毁用户连接
const CMD_DESTROY = Protocol(8)

// 发给gateway，通知用户session更新
const CMD_UPDATE_SESSION = Protocol(9)

// 获取在线状态
const CMD_GET_ALL_CLIENT_SESSIONS = Protocol(10)

// 判断是否在线
const CMD_IS_ONLINE = Protocol(11)

// client_id绑定到uid
const CMD_BIND_UID = Protocol(12)

// 解绑
const CMD_UNBIND_UID = Protocol(13)

// 向uid发送数据
const CMD_SEND_TO_UID = Protocol(14)

// 根据uid获取绑定的client_id
const CMD_GET_CLIENT_ID_BY_UID = Protocol(15)

// 加入组
const CMD_JOIN_GROUP = Protocol(20)

// 离开组
const CMD_LEAVE_GROUP = Protocol(21)

// 向组成员发消息
const CMD_SEND_TO_GROUP = Protocol(22)

// 获取组成员
const CMD_GET_CLIENT_SESSIONS_BY_GROUP = Protocol(23)

// 获取组在线连接数
const CMD_GET_CLIENT_COUNT_BY_GROUP = Protocol(24)

// 按照条件查找
const CMD_SELECT = Protocol(25)

// 获取在线的群组ID
const CMD_GET_GROUP_ID_LIST = Protocol(26)

// 取消分组
const CMD_UNGROUP = Protocol(27)

// worker连接gateway事件
const CMD_WORKER_CONNECT = Protocol(200)

// 心跳
const CMD_PING = Protocol(201)

// GatewayClient连接gateway事件
const CMD_GATEWAY_CLIENT_CONNECT = Protocol(202)

// 根据client_id获取session
const CMD_GET_SESSION_BY_CLIENT_ID = Protocol(203)

// 发给gateway，覆盖session
const CMD_SET_SESSION = Protocol(204)

// 当websocket握手时触发，只有websocket协议支持此命令字
const CMD_ON_WEBSOCKET_CONNECT = Protocol(205)

// 包体是标量
const FLAG_BODY_IS_SCALAR = Protocol(0x01)

// 通知gateway在send时不调用协议encode方法，在广播组播时提升性能
const FLAG_NOT_CALL_ENCODE = Protocol(0x02)

// 包头长度
const HEAD_LEN = 28

func Input(buffer []byte) int {
	if len(buffer) < HEAD_LEN {
		return 0
	}

	return len(buffer)
}
