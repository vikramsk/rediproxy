package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

var (
	redisURL = flag.String("redis-url", "localhost:6379", "URL for Redis")
	proxyURL = flag.String("proxy-url", "localhost:8080", "service URL for rediproxy")

	// serviceURL is the endpoint for rediproxy.
	serviceURL string

	// testPrefix is a random number that's prefixed to each key for the
	// given test run. This is added in order to minimize the odds of finding a
	// cache hit for a key that was added for a previous run for the tests.
	testPrefix int
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
	testPrefix = rand.Int()
	serviceURL = fmt.Sprintf("http://%s/cache?key=", *proxyURL)
}

func key(i int) string {
	return fmt.Sprintf("e2e%dKey%d", testPrefix, i)
}

func value(i int) string {
	return fmt.Sprintf("e2e%dValue%d", testPrefix, i)
}

// testRun encapsulates a single key fetch request.
type testRun struct {

	// key represents the input for the request.
	key int

	// response is the HTTP response returned for the
	// requested key.
	response *http.Response
}

func createRedisClient(addr string) (*redis.Client, error) {
	redisdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := redisdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	return redisdb, nil
}

func TestE2EFlow(t *testing.T) {
	c, err := createRedisClient(*redisURL)
	if err != nil {
		t.Fatalf("could not create client, err: %v", err)
	}
	keyLimit := 1024
	for i := 0; i < keyLimit; i++ {
		c.Del(key(i))
	}
	for i := 0; i < keyLimit; i++ {
		c.Set(key(i), value(i), 0)
	}

	responseChan := make(chan *testRun, keyLimit*2)
	var wg sync.WaitGroup
	for i := 0; i < keyLimit*2; i++ {
		wg.Add(1)
		go callRediproxy(t, &wg, rand.Int()%keyLimit*2, responseChan)
	}
	wg.Wait()
	close(responseChan)
	for r := range responseChan {
		assertResponse(t, keyLimit, r)
	}
}

// assertResponse performs the assertions on the response values for a test run.
func assertResponse(t *testing.T, keyLimit int, s *testRun) {
	if s.key >= keyLimit && s.response.StatusCode != http.StatusNoContent {
		t.Errorf("expected StatusCode %d for key: %s, but received %d", http.StatusNoContent, key(s.key), s.response.StatusCode)
	}

	body, _ := ioutil.ReadAll(s.response.Body)
	defer s.response.Body.Close()
	if s.key < keyLimit && (string(body) != value(s.key) || s.response.StatusCode != http.StatusOK) {
		t.Errorf("key should be found in cache with Status OK. key: %d", s.key)
	}
}

func callRediproxy(t *testing.T, wg *sync.WaitGroup, k int, respChan chan<- *testRun) {
	defer wg.Done()
	resp, err := http.Get(serviceURL + key(k))
	if err != nil {
		t.Fatalf("unexpected error while connecting to rediproxy. err: %v", err)
	}
	respChan <- &testRun{
		key:      k,
		response: resp,
	}
}
