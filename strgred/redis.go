package strgred

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// структура для подключения к redis
type Config struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

// Функция подключения к redis
func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	// проверка соединения (пингуем по пустому контексту)
	if err := db.Ping(ctx).Err(); err != nil {
		log.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}

// конкретный конфиг подключения к бд
var cfg Config = Config{
	Addr:        "localhost:6380",
	Password:    "ylp3QnB(VR0v>oL<Y3heVgsdE)+O+RZ",
	User:        "leosah",
	DB:          0,
	MaxRetries:  5,
	DialTimeout: 10 * time.Second,
	Timeout:     5 * time.Second,
}

func Redis_add(key int64, status, acсess_token, update_token string) {

	db, err := NewClient(context.Background(), cfg)
	if err != nil {
		log.Panic("db creating fail: ", err)
	}
	// вносим значение
	value := CreateValue3(status, acсess_token, update_token)
	err2 := db.Set(context.Background(), fmt.Sprint(key), value, 0).Err()
	if err2 != nil {
		log.Panic("ERROR in redis_add: ", err)
	}
}

func Redis_add2(key int64, status, entry_token string) {

	db, err := NewClient(context.Background(), cfg)
	if err != nil {
		log.Panic("db creating fail: ", err)
	}
	// вносим значение
	value := CreateValue2(status, entry_token)
	err2 := db.Set(context.Background(), fmt.Sprint(key), value, 0).Err()
	if err2 != nil {
		log.Panic("ERROR in redis_add: ", err)
	}
}

func Redis_get(key any) (string, string, string) {

	db, err := NewClient(context.Background(), cfg)
	if err != nil {
		log.Panic("db creating fail: ", err)
	}
	// получаем значение из бд
	value, err := db.Get(context.Background(), fmt.Sprint(key)).Result()
	if err == redis.Nil {
		return "nil", "", ""
	}
	return SplitValue(value)
}

func Redis_delete(key any) bool {

	db, err := NewClient(context.Background(), cfg)
	if err != nil {
		log.Panic("db creating fail: ", err)
	}

	ok, err := db.Del(context.Background(), fmt.Sprint(key)).Result()
	if err != nil {
		log.Panic("ERROR in redis_delete: ", err)
	}

	return ok > 0
}

// поиск ключей по значению find
func GetSomeIDs(find string) []int64 {

	db, err := NewClient(context.Background(), cfg)
	if err != nil {
		log.Panic("db creating fail: ", err)
	}

	var results []int64

	keys, err2 := db.Keys(context.Background(), "*").Result()
	if err2 != nil {
		log.Panic("ERROR in GetSomeIDs: ", err2)
	}
	//log.Println(keys)

	for _, key := range keys {
		status, _, _ := Redis_get(key)
		intkey, errstr := strconv.Atoi(key)
		if errstr != nil {
			log.Panic("ERROR in GetSomeIDs: strconv error: ", errstr)
		}
		if status == find {
			results = append(results, int64(intkey))
		}
	}
	return results
}

// объединяем статус, access_token и update_token
func CreateValue3(status, access_token, update_token string) string {
	return status + "\n" + access_token + "\n" + update_token
}

// объединаем статус и entry_token
func CreateValue2(status, entry_tocken string) string {
	return status + "\n" + entry_tocken
}

// достаём из строки разные штуки
func SplitValue(value string) (string, string, string) {
	lines := strings.SplitAfter(value, "\n")
	status := lines[0]
	var token1, token2 string
	token1 = lines[1] //тут либо токен входа (тогда второй пустой) либо токен доступа (тогда второй токен обновления)

	if len(lines) > 2 {
		token2 = lines[2] //токен обновления
	} else {
		token2 = ""
	}
	return status, token1, token2
}

func GenerateEntryToken() string {
	ABC := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890.<>_-=+*[]{}()"
	result := ""
	for i := 0; i < 16; i++ {
		rand.Seed(time.Now().UnixNano())
		result += string(ABC[rand.Intn(76)])
	}
	return result
}