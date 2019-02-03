// subscale fixes up mis-timed srt files by translating and scaling.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type tm int

const (
	millisecond tm = 1
	second         = 1000 * millisecond
	minute         = 60 * second
	hour           = 60 * minute
)

var oldzero, newzero tm // translation
var scale = 1.0

func transtm(t tm) tm {
	t = newzero + tm(float64(t-oldzero)*scale)
	if t < 0 {
		return 0
	}
	return t
}

func parsetm(ts string) (tm, error) {
	var h, m, s, ms tm
	_, err := fmt.Sscanf(ts, "%d:%d:%d,%d", &h, &m, &s, &ms)
	if err != nil {
		// try period separator
		_, err = fmt.Sscanf(ts, "%d:%d:%d.%d", &h, &m, &s, &ms)
	}
	return h*hour + m*minute + s*second + ms*millisecond, err
}

func mustparsetm(t string) tm {
	d, e := parsetm(t)
	if e != nil {
		fatal("could not parse time:", t, e)
	}
	return d
}

func formattm(t tm) string {
	h := t / hour
	m := (t % hour) / minute
	s := t % minute / second
	ms := t % second / millisecond
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func transln(s string) string {
	ab := strings.Split(s, " --> ")
	if len(ab) != 2 {
		return s
	}
	a, aerr := parsetm(ab[0])
	b, berr := parsetm(ab[1])
	if aerr != nil || berr != nil {
		// normal subtitle that happend to contain " --> ", presumably

		// log.Fatal("could not offset", a, aerr, b, berr)
		return s
	}
	return formattm(transtm(a)) + " --> " + formattm(transtm(b))
}

func main() {
	if (len(os.Args) != 3 && len(os.Args) != 5) || os.Args[1] == "-h" {
		fmt.Fprintf(os.Stderr, `usage: %s OLD_TIME NEW_TIME [OLD2 NEW2] < OLD_SRT > NEW_SRT	

Fix subtitle timing in a .srt file.

Example:
	%s 00:02:35,300 00:3:36,700 < old.srt > new.srt
`, os.Args[0], os.Args[0])
		os.Exit(0)
	}
	oldzero, newzero = mustparsetm(os.Args[1]), mustparsetm(os.Args[2])
	if len(os.Args) == 5 {
		ot, nt := mustparsetm(os.Args[3]), mustparsetm(os.Args[4])
		if ot == oldzero {
			fatal("error: same time specificed twice", os.Args[2])
		}
		scale = float64(nt-newzero) / float64(ot-oldzero)
	}
	f := bufio.NewScanner(os.Stdin)
	for f.Scan() {
		fmt.Println(transln(f.Text()))
	}
}

func fatal(m ...interface{}) {
	fmt.Fprintln(os.Stderr, m...)
	os.Exit(1)
}
