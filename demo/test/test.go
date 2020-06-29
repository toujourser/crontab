package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

func main() {
	var (
		exe          string
		err          error
		filedir      string
		fileLock     *flock.Flock
		lockFilePath string
		locked       bool
	)
	if exe, err = os.Executable(); err != nil {
		log.Println(err.Error())
	}
	if filedir, err = filepath.Abs(filepath.Dir(exe)); err != nil {
		log.Println(err.Error())
	}

	fmt.Printf("%+v\n", exe)
	fmt.Printf("%+v\n", filedir)
	lockFilePath = filepath.Join(filedir, "bin", "test.lock")
	fileLock = flock.New(lockFilePath)

	if locked, err = fileLock.TryLock(); err != nil {
		log.Println("lock failed ....")
	}
	defer func() {
		fileLock.Unlock()
		if err = os.Remove(lockFilePath); err != nil {
			log.Fatal(err.Error())
		}
	}()
	fmt.Printf("%+v %+v\n", locked, err)
	if !locked {
		log.Println("lock failed")
		return
	}

	time.Sleep(10 * time.Second)

}
