include $(GOROOT)/src/Make.inc

LDFLAGS=$(shell pkg-config oniguruma --libs)
CFLAGS=$(shell pkg-config oniguruma --cflags)

TARG=rubex

CGOFILES=\
  regex.go\

GOFILES=\
  constants.go\

CGO_OFILES=\
  chelper.o\

CLEANFILES+=

include $(GOROOT)/src/Make.pkg

%.o: %.c
	gcc $(_CGO_CFLAGS_$(GOARCH)) -g -O2 -fPIC $(CFLAGS) -o $@ -c $^

onig:
	cd ./oniguruma && ./configure && make && sudo make install

onig-clean:
	cd ./oniguruma && make distclean

