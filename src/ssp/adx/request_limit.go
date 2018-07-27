package adx

import (
	"fmt"
	"ssp/util"
	"time"
	"tripod/devkit"

	"github.com/garyburd/redigo/redis"
)

const imLuaScript = `
if redis.call('TTL', KEYS[1]) < 0 then
	redis.call('HMSET', KEYS[1], ARGV[1], 1, ARGV[2], 1);
	redis.call('EXPIREAT', KEYS[1], ARGV[4]);
else 
	redis.call('HINCRBY', KEYS[1], ARGV[1], 1);
	if redis.call('HEXISTS', KEYS[1], ARGV[2]) == 1 then
		redis.call('HINCRBY', KEYS[1], ARGV[2], 1);
	else
		redis.call('HMSET', KEYS[1], ARGV[2], 1);
		redis.call('HDEL', KEYS[1], ARGV[3]);
	end;
end `

var (
	adxRedisPool = util.AdxRedisPool
	imScript     *redis.Script
)

const (
	preReq              = "req"
	preImp              = "imp"
	smoothScale float64 = 0.05
)

func init() {
	imScript = redis.NewScript(1, imLuaScript)
}

func makeKey(adslotId string) string {
	return fmt.Printf("adslot_%s", adslotId)
}

func makeField(prefix string, d int) string {
	now := time.Now()
	date := now.AddDate(0, 0, d)
	dateStr := date.Format("20060102")
	return fmt.Printf("%s_%s", prefix, dateStr)
}

func makeFieldTotal(prefix string) string {
	return fmt.Printf("%s_total", prefix)
}

func GetAdslotInfo(adslotId string) map[string]int {
	// Ad slot info contains: req_total, req_<day>, imp_total, imp_<day>
	adslotInfo = make(map[string]int)
	key := makeKey(adslotId)
	conn := adxRedisPool.GetConn(key)
	defer conn.Close()

	result, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		// Receive request from this adslot first time, its key hasnt been set
		// in redis, will return a redis.ErrNil
		util.Log.Error("GetAdslotInfo for %s:%s", key, err.Error())
		return nil
	}
	if 8 != len(result) {
		util.Log.Error("Ad slot key:%s, field num:%d", key, len(result))
		return nil
	}
	for i := 0; i < len(result); {
		i = i + 2
		if i+1 < len(result) {
			k, err := redis.String(result[i], nil)
			if err != nil {
				util.Log.Error("Get adslot info field key:%s", err.Error())
				return nil
			}
			v, err := redis.Int64(result[i], nil)
			if err != nil {
				util.Log.Error("Get adslot info field value:%s", err.Error())
				return nil
			}
			adslotInfo[k] = v
		}
	}

	return adslotInfo
}

func IncField(adslotId, prefix string, expireAt string) error {
	key := makeKey(adslotId)
	fieldToday := makeField(prefix, 0)
	fieldYesterday := makeField(prefix, -1)
	filedTotal := makeFieldTotal(prefix)
	conn := adxRedisPool.GetConn(key)
	timeLayout := "2006-01-02"
	loc, _ := time.LoadLocation("Local")
	expireTime, _ := time.ParseInLocation(timeLayout, expireAt, loc)
	timestamp := expireTime.Unix()

	defer conn.Close()
	_, err := imScript.Do(conn, key, filedTotal, fieldToday, fieldYesterday, timestamp)
	return err
}

func checkReq(adslotId string, adslotInfo map[string]int, slotConfig *util.SlotConfig) bool {
	if nil == slotConfig {
		util.Log.Error("The slotConfig is nil, adslot id:%s", adslotId)
		return false
	}
	if nil == adslotInfo {
		return true
	}
	if reqTotal := adslotInfo[makeFieldTotal(preReq)]; reqTotal > slotConfig.RequestTotal {
		util.Log.Debug("Request total is over limit, adslot id:%s, req total: %d",
			adslotId, reqTotal)
		return false
	}
	if reqDaily := adslotInfo[makeField(preReq, 0)]; reqDaily > slotConfig.RequestDaily {
		util.Log.Debug("Request daily is over limit, adslot id:%s, req daily:%d",
			adslotId, reqDaily)
		return false
	}
	if impTotal := adslotInfo[makeFieldTotal(preImp)]; impTotal > slotConfig.ImpressionTotal {
		util.Log.Debug("Impression total is over limit, adslot id:%s, imp total:%d",
			adslotId, impTotal)
		return false
	}
	if impDaily := adslotInfo[makeField(preImp, 0)]; impDaily > slotConfig.ImpressionDaily {
		util.Log.Debug("Impression daily is over limit, adslot id:%s, imp daily:%d",
			adslotId, impDaily)
		return false
	}
	IncField(adslotId, preReq, slotConfig.EndDate)
	return true
}

func smoothControl(adslotId string, adslotInfo map[string]int, slotConfig *util.SlotConfig) bool {
	// Only control request smoothly
	if nil == adslotInfo {
		return true
	}
	year, month, day := time.Now().Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	t := time.Now()
	percent := float64(t.Unix()-today.Unix()) / float64(86400)
	target := percent * (float64(slotConfig.RequestDaily)) * (1 + smoothScale)
	if reqDaily := adslotInfo[makeField(preReq, 0)]; reqDaily > target {
		util.Log.Debug("Smooth control, adslot id:%s, req daily now:%d, target:%d",
			adslotId, reqDaily, target)
		return false
	}

	return true
}
