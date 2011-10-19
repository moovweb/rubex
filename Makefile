include $(GOROOT)/src/Make.inc

LDFLAGS=-L/usr/local/lib -lonig
CFLAGS=-I/usr/local/include

TARG=rubex

CGOFILES=\
  cgoflags.go\
  regex.go\

GOFILES=\
  constants.go\

CGO_OFILES=\
  chelper.o\

CLEANFILES+=

include $(GOROOT)/src/Make.pkg

%.o: %.c
	gcc $(_CGO_CFLAGS_$(GOARCH)) -g -O2 -fPIC $(CFLAGS) -o $@ -c $^
