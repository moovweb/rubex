// Comparing the speeds of the golang native regex library and rubex.
// The numbers show a dramatic difference, with rubex being nearly 400 
// times slower than the native go libraries.  Unfortunately for us,
// the native go libraries have a different regex behavior than rubex,
// so we'll have to hack at it a bit to fit our needs if we decide to use it.
// (which we should, I mean, come on, 400 times faster?  That's mad wins.)

package main

import re "rubex"
import "time"
import "regexp"
import "runtime"
import "os"
import "strconv"

var re1 []Matcher
var re2 []Matcher
const NUM = 100
const NNN = 1000
var STR = "abcdabc"

type Matcher interface {
	MatchString(string) bool
}

type Task struct {
	str string
	m Matcher
	t time.Time
}

var TaskChann chan *Task

func init() {
	re1 = make([]Matcher, NUM)
	re2 = make([]Matcher, NUM)
	for i := 0; i < NUM; i ++ {
		re1[i] = regexp.MustCompile("[a-c]*$")
		re2[i] = re.MustCompile("[a-c]*$")
	}
	TaskChann = make(chan *Task, 100)
	for i := 0; i < 10; i ++ {
		STR += STR
	}
	println("len:", len(STR))
}

func render_pages(name string, marray []Matcher, num_routines, num_renders int) {
	for i := 0; i < num_routines; i++ {
		m := marray[i]
		go func () {
			runtime.LockOSThread()
			for j := 0; j < num_renders; j++ {
				var totalDuration int64 = 0
				for i := 0; i < NNN; i++ {
					t := time.Now()
					m.MatchString(STR)
					totalDuration += time.Since(t).Nanoseconds()
				}
				println(name + "-average: ",  totalDuration/int64(1000*NNN), "us")
			}
		}()
	}
}

func render_pages2(name string, marray []Matcher, num_routines, num_renders int) {
	go func() {
		for i := 0; i < 100000; i ++ {
			t := &Task{str: STR, m: marray[0], t: time.Now()}
			TaskChann <- t
		}
	}()
	for i := 0; i < num_routines; i++ {
		m := marray[i]
		go func () {
			runtime.LockOSThread()
			for j := 0; j < num_renders; j++ {
				var totalDuration int64 = 0
				for i := 0; i < NNN; i++ {
					task := <-TaskChann
					m.MatchString(task.str)
					totalDuration += time.Since(task.t).Nanoseconds()
				}
				println(name + "-average: ",  totalDuration/int64(1000*NNN), "us")
			}
		}()
	}
}



func main() {
	cpu, _ := strconv.Atoi(os.Args[1])
	lib := os.Args[2]
	println("using CPUs:", cpu)
	runtime.GOMAXPROCS(cpu)
	num_routines := 2
	num_renders := 20
	
	if lib == "rubex" {
		render_pages2("rubex", re2, num_routines, num_renders)
	} else {
		render_pages2("regexp", re1, num_routines, num_renders)
	}

	d, _ := time.ParseDuration("5s")
	for i := 0; i < 100; i ++ {
		println("goroutine:", runtime.NumGoroutine())
		time.Sleep(d)
		
	}
	println ("Done")
}

