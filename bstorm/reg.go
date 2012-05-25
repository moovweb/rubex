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

var re1 []*regexp.Regexp
var re2 []*re.Regexp
const NUM = 100
var STR = "abcdabc"

func init() {
	re1 = make([]*regexp.Regexp, NUM)
	re2 = make([]*re.Regexp, NUM)
	for i := 0; i < NUM; i ++ {
		re1[i] = regexp.MustCompile("[a-c]*$")
		re2[i] = re.MustCompile("[a-c]*$")
	}
	for i := 0; i < 10; i ++ {
		STR += STR
	}
	println("len:", len(STR))
}

func rubex_page(index int) {
	for i := 0; i < 100; i++ {
		r := re2[index]
		r.MatchString(STR)
	}
}

func regexp_page(index int) {
	for i := 0; i < 100; i++ {
		r := re1[index]
		r.MatchString(STR)
	}
}

func render_pages(name string, fp  func(int), num_routines, num_renders int) {
	for i := 0; i < num_routines; i++ {
		go func () {
			t := time.Now()
			for j := 0; j < num_renders; j++ {
				fp(i)	
			}
			println(name + "-average: ",  time.Since(t).Nanoseconds()/int64(num_renders*1000000), "ms")
		}()
	}
}

func main() {
	cpu, _ := strconv.Atoi(os.Args[1])
	lib := os.Args[2]
	println("using CPUs:", cpu)
	runtime.GOMAXPROCS(cpu)
	num_routines := 90
	num_renders := 20
	
	if lib == "rubex" {
		render_pages("rubex", rubex_page, num_routines, num_renders)
	} else {
		render_pages("regexp", regexp_page, num_routines, num_renders)
	}

	d, _ := time.ParseDuration("5s")
	for i := 0; i < 100; i ++ {
		println("goroutine:", runtime.NumGoroutine())
		time.Sleep(d)
		
	}
	println ("Done")
}

