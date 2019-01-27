package main

import (
	"context"
	"github.com/go-locks/distlock"
	"github.com/go-locks/distlock/mutex"
	"github.com/go-locks/redis-driver"
	"github.com/letsfire/redigo"
	"github.com/letsfire/redigo/mode/alone"
	"log"
	"time"
)

func main() {
	/* redis deploy mode, one of [alone,cluster,sentinel],
	 * more usage see https://github.com/letsfire/redigo. */
	md := alone.New(alone.Addr("192.168.0.110:6379"))

	/* instantiate a redigo instance by your mode. */
	rg := redigo.New(md)

	/* now you can instantiate redis driver for distributed locks use redigo instance.
	 * you can specify multiple redigo instances to protect from single node of failure. */
	rd := redis.New(rg /* more redigo instances here if you need. */)

	/* instantiate a distlock instance by your redis driver. */
	dl := distlock.New(rd)

	/* create a mutex named `demo` and a rwmutex named `demo.rw` and specify their deadline and factor,
	 * more usage see https://github.com/go-locks/distlock. */
	mtx, _ := dl.NewMutex("demo", mutex.Expiry(time.Second*2), mutex.Factor(0.30))
	rwMtx, _ := dl.NewRWMutex("demo.rw", mutex.Expiry(time.Second*2), mutex.Factor(0.30))

	lockDemo(mtx, "lock mutex")                 // Blocking until acquire lock success
	lockCtxDemo(mtx, "lock mutex with context") // Blocking until acquire lock success (true) or ctx.Done (false)
	tryLockDemo(mtx, "try lock mutex")          // Non-Blocking, return true if acquire lock success, otherwise return false

	/* rwmutex separate into read mutex and write mutex, their have the same usage with mutex */
	rMtx := rwMtx.Read()
	wMtx := rwMtx.Write()

	lockDemo(rMtx, "lock read mutex")
	lockCtxDemo(rMtx, "lock read mutex with context")
	tryLockDemo(rMtx, "try lock read mutex")

	lockDemo(wMtx, "lock write mutex")
	lockCtxDemo(wMtx, "lock write mutex with context")
	tryLockDemo(wMtx, "try lock write mutex")
}

func lockDemo(mtx *mutex.Mutex, msg string) {
	mtx.Lock()

	// the processing time must be guaranteed to be shorter than the deadline of mutex.
	// if you cannot guarantee it，you should better to touch the mutex before overdue.
	ctx, cancel := context.WithCancel(context.TODO())
	mtx.Heartbeat(ctx)

	// handle your business logic here.
	time.Sleep(time.Second * 2)
	log.Println(msg)
	cancel()

	mtx.Unlock()
}

func lockCtxDemo(mtx *mutex.Mutex, msg string) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second)
	if mtx.LockCtx(ctx) {
		// the processing time must be guaranteed to be shorter than the deadline of mutex.
		// if you cannot guarantee it，you should better to touch the mutex before overdue.
		ctx, cancel := context.WithCancel(context.TODO())
		mtx.Heartbeat(ctx)

		// handle your business logic here.
		time.Sleep(time.Second * 2)
		log.Println(msg)
		cancel()

		mtx.Unlock()
	}
}

func tryLockDemo(mtx *mutex.Mutex, msg string) {
	if mtx.TryLock() {
		// the processing time must be guaranteed to be shorter than the deadline of mutex.
		// if you cannot guarantee it，you should better to touch the mutex before overdue.
		ctx, cancel := context.WithCancel(context.TODO())
		mtx.Heartbeat(ctx)

		// handle your business logic here.
		time.Sleep(time.Second * 2)
		log.Println(msg)
		cancel()

		mtx.Unlock()
	}
}
