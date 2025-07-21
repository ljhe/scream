package redis

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client              *redis.Client
	defaultClientConfig = &Parm{
		redisopt: redis.Options{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			MinIdleConns: 128,
			PoolSize:     1024,
		},
	}
)

func BuildClientWithOption(opts ...Option) *redis.Client {

	for _, opt := range opts {
		opt(defaultClientConfig)
	}
	return new(defaultClientConfig)
}

func new(p *Parm) *redis.Client {
	// 创建连接池
	client = redis.NewClient(&p.redisopt)
	ctx := context.Background()
	//判断是否能够链接到数据库

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	return client
}

func GetClient() *redis.Client {
	return client
}

func MockClient(cli *redis.Client) {
	client = cli
}

// Pipeline 管道
func Pipeline() redis.Pipeliner {
	return client.Pipeline()
}

type PipelineFunc func(pipe redis.Pipeliner) error

// TxPipelined 封装事务,将命令包装在MULTI、EXEC中,并直接执行事务
func TxPipelined(ctx context.Context, from string, fn PipelineFunc) ([]redis.Cmder, error) {
	cmds, err := client.TxPipelined(ctx, fn)
	return cmds, err
}

func Pipelined(ctx context.Context, from string, fn PipelineFunc) ([]redis.Cmder, error) {
	return client.Pipelined(ctx, fn)
}

func ConnHGet(conn *redis.Conn, ctx context.Context, key, field string) (string, error) {
	cmd := conn.HGet(context.TODO(), key, field)
	return cmd.Val(), cmd.Err()
}

// Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
func Get(ctx context.Context, key string) *redis.StringCmd {
	return client.Get(ctx, key)
}

func Set(ctx context.Context, key string, val string) *redis.StatusCmd {
	return client.Set(ctx, key, val, 0)
}

func Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return client.Del(ctx, keys...)
}

// SetEx Redis `SETEx key expiration value` command.
func SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return client.SetEx(ctx, key, value, expiration)
}

func SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return client.SetNX(ctx, key, value, expiration)
}

// ------------ Set 集合 ------------

func SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return client.SAdd(ctx, key, members...)
}

func SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return client.SIsMember(ctx, key, member)
}

func SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return client.SMembers(ctx, key)
}

func SRandMember(ctx context.Context, key string) *redis.StringCmd {
	return client.SRandMember(ctx, key)
}

func SRandMemberN(ctx context.Context, key string, count int64) *redis.StringSliceCmd {
	return client.SRandMemberN(ctx, key, count)
}

func SPop(ctx context.Context, key string) *redis.StringCmd {
	return client.SPop(ctx, key)
}

func SCard(ctx context.Context, key string) *redis.IntCmd {
	return client.SCard(ctx, key)
}

// ------------ ZSet 有序集合 ----------

func ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	return client.ZScore(ctx, key, member)
}

func ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return client.ZAdd(ctx, key, members...)
}

func ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	return client.ZRevRank(ctx, key, member)
}

func ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return client.ZRevRange(ctx, key, start, stop)
}

func ZCard(ctx context.Context, key string) *redis.IntCmd {
	return client.ZCard(ctx, key)
}

func ZRangeByScore(ctx context.Context, key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	return client.ZRangeByScore(ctx, key, &opt)
}

func ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return client.ZRem(ctx, key, members...)
}

// ----------- Hash 哈希 ---------------

func HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return client.HGet(ctx, key, field)
}

func HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return client.HGetAll(ctx, key)
}

func HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return client.HSet(ctx, key, values...)
}

func HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return client.HDel(ctx, key, fields...)
}

func HLen(ctx context.Context, key string) *redis.IntCmd {
	return client.HLen(ctx, key)
}

func HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd {
	return client.HIncrBy(ctx, key, field, incr)
}

func HKeys(ctx context.Context, key string) *redis.StringSliceCmd {
	return client.HKeys(ctx, key)
}

func Exists(ctx context.Context, key string) *redis.IntCmd {
	return client.Exists(ctx, key)
}

func HExists(ctx context.Context, key, field string) *redis.BoolCmd {
	return client.HExists(ctx, key, field)
}

func CAD(ctx context.Context, key, field string) bool {
	ret, err := client.Do(ctx, "CAD", key, field).Bool()
	if err != nil {
		fmt.Println("cad err", err.Error())
	}
	return ret
}

func XInfoGroups(ctx context.Context, key string) *redis.XInfoGroupsCmd {
	return client.XInfoGroups(ctx, key)
}

func XGroupCreate(ctx context.Context, key, group, start string) *redis.StatusCmd {
	return client.XGroupCreate(ctx, key, group, start)
}

func XGroupDelConsumer(ctx context.Context, key, group, consumer string) *redis.IntCmd {
	return client.XGroupDelConsumer(ctx, key, group, consumer)
}

func XInfoConsumers(ctx context.Context, key, group string) *redis.XInfoConsumersCmd {
	return client.XInfoConsumers(ctx, key, group)
}

func XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	return client.XReadGroup(ctx, a)
}

func XGroupDestroy(ctx context.Context, key, group string) *redis.IntCmd {
	return client.XGroupDestroy(ctx, key, group)
}

func XLen(ctx context.Context, key string) *redis.IntCmd {
	return client.XLen(ctx, key)
}

func XDel(ctx context.Context, key, id string) *redis.IntCmd {
	return client.XDel(ctx, key, id)
}

func ScriptRun(ctx context.Context, script *redis.Script, keys []string, args ...any) (interface{}, error) {
	val, err := script.Run(ctx, client, keys, args...).Result()
	return val, err
}

func ScriptRunInt64s(ctx context.Context, script *redis.Script, keys []string, args ...any) ([]int64, error) {
	val, err := script.Run(ctx, client, keys, args...).Int64Slice()
	return val, err
}

func LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return client.LRange(ctx, key, start, stop)
}

func Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return client.Expire(ctx, key, expiration)
}

func RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return client.RPush(ctx, key, values...)
}

func FlushDB(ctx context.Context) *redis.StatusCmd {
	return client.FlushDB(ctx)
}

func FlushAll(ctx context.Context) *redis.StatusCmd {
	return client.FlushAll(ctx)
}

func PoolStats(ctx context.Context) *redis.PoolStats {
	return client.PoolStats()
}

func XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	return client.XAdd(ctx, a)
}

func Incr(ctx context.Context, key string) *redis.IntCmd {
	return client.Incr(ctx, key)
}

func Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	return client.Keys(ctx, pattern)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
