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


func rubex_page() {
	for i := 0; i < 100; i++ {
		r := re.MustCompile("[a-c]*$")
		r.MatchString("abcdabc")
		r.Free()
	}
}

func regexp_page() {
	for i := 0; i < 100; i++ {
		r := regexp.MustCompilePOSIX("[a-c]*$")
		r.MatchString("abcdabc")
		//r.Free()
	}
}

func render_pages(name string, fp  func(), num_routines, num_renders int) {
	for i := 0; i < num_routines; i++ {
		go func () {
			for j := 0; j < num_renders; j++ {
				t := time.Now()
				fp()	
				println(name + "-average: " + time.Since(t).String())
			}
		}()
	}
}

func main() {

	num_routines := 40
	num_renders := 100
	
//	render_pages("rubex", rubex_page, num_routines, num_renders)
	render_pages("regexp", regexp_page, num_routines, num_renders)

	d, _ := time.ParseDuration("5m")
	time.Sleep(d)
	println ("Done")
}

