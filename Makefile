include $(GOROOT)/src/Make.inc

LDFLAGS=-L/usr/local/lib -lonig
CFLAGS=-I/usr/local/include

TARG=rubex

CGOFILES=\
  cgoflags.go\
  regex.go\

GOFILES=\
  constants.go\
  helper.go

CGO_OFILES=\
  chelper.o\

CLEANFILES+=

include $(GOROOT)/src/Make.pkg
