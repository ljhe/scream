package utils

import "time"

const DateOnly = "20060102"

func GetDateOnly() string {
	return time.Now().Format(DateOnly)
}

func GetDateTime() string {
	return time.Now().Format(time.DateTime)
}

func GetLoc() *time.Location {
	//if gameLoc == nil {
	//	//loc, err := time.LoadLocation("Asia/Shanghai")
	//	loc, err := time.LoadLocation("Asia/Tokyo")
	//	if err != nil {
	//		gameLoc = time.Local
	//	} else {
	//		gameLoc = loc
	//	}
	//}
	//return gameLoc
	return time.Local // 使用系统时区
}

func GetCurrentTimeMs() uint64 {
	t1 := GetCurrentTimeNow()
	return uint64(t1.UnixNano() / 1e6)
}

func GetCurrentTimeNow() time.Time {
	loc := GetLoc()
	t1 := time.Now()
	return t1.In(loc)
}

func GetTimeSeconds() int64 {
	return GetCurrentTimeNow().Unix()
}
