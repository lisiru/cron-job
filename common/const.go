package common

const (
	JOB_STAT_READY int = iota + 1
	JOB_STAT_DELAY
	JOB_STAT_RESERVED
	JOB_STAT_DELETED
)

const (
	JOB_INFO_HASH_KEY_PREFIX = "job_info_hash_key:"   // 后面接Jobid
	JOB_INFO_PREFIX          = "job_info:"            // 后面接JobId
	DELAY_JOB_BUKET_ZSET_KEY = "delay_job_buket_zset" // 延迟任务的池，存储结构为redis 的zset
	JOB_READY_QUEUE_KEY      = "job_readey_queue"     // 可以执行的任务队列
	JOB_LOCK_KEY = "job_lock_key"// 加锁的key

)

const (
	STRS        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	LOCK_COMMAND = `if redis.call("GET",KEYS[1]) == ARGV[1] then redis.call("SET",KEYS[1],ARGV[1],"PX",ARGV[2]) return "OK" else return redis.call("SET",KEYS[1],ARGV[1],"NX","PX",ARGV[2]) end`
	DEL_COMMAND =`if redis.call("GET",KEYS[1]) == ARGV[1] then return redis.call("DEL",KEYS[1]) else return 0 end`

	RANDOM_STR_LEN = 16

	// 默认超时时间，防止死锁
	EXPIRE = 500
)
