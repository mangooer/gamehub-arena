package cache

import "fmt"

const (
	// 用户相关键
	KeyUserSession = "user:session:%d" // 用户会话
	KeyUserInfo    = "user:info:%d"    // 用户信息
	KeyUserStatus  = "user:status:%d"  // 用户状态
	KeyUsersOnline = "users:online"    // 在线用户集合

	// 游戏相关键
	KeyGameRoom    = "room:%s"         // 游戏房间
	KeyRoomPlayers = "room:%s:players" // 房间玩家
	KeyRoomQueue   = "room:queue"      // 房间队列

	// 匹配相关键
	KeyMatchQueue   = "match:queue:%s"   // 匹配队列
	KeyMatchHistory = "match:history:%d" // 匹配历史

	// 排行榜相关键
	KeyLeaderboard    = "leaderboard:%s"     // 排行榜
	KeyLeaderboardTTL = "leaderboard:ttl:%s" // 排行榜TTL

	// 统计相关键
	KeyStatsDaily  = "stats:daily:%s"  // 每日统计
	KeyStatsHourly = "stats:hourly:%s" // 每小时统计

	// 锁相关键
	KeyLockUser  = "lock:user:%d"  // 用户锁
	KeyLockRoom  = "lock:room:%s"  // 房间锁
	KeyLockMatch = "lock:match:%d" // 匹配锁
)

// 生成键的辅助函数
func UserSessionKey(userID uint64) string {
	return fmt.Sprintf(KeyUserSession, userID)
}

func UsersOnlineKey() string {
	return KeyUsersOnline
}

func UserStatusKey(userID uint64) string {
	return fmt.Sprintf(KeyUserStatus, userID)
}

func UserInfoKey(userID uint64) string {
	return fmt.Sprintf(KeyUserInfo, userID)
}

func GameRoomKey(roomCode string) string {
	return fmt.Sprintf(KeyGameRoom, roomCode)
}

func MatchQueueKey(gameMode string) string {
	return fmt.Sprintf(KeyMatchQueue, gameMode)
}

func LeaderboardKey(leaderboardType string) string {
	return fmt.Sprintf(KeyLeaderboard, leaderboardType)
}
