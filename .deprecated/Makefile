include $(GOROOT)/src/Make.inc

TARG=rubex
CGOFILES=regex.go
#GOFILES=regex.go 
ONIG_CONFIG=$(shell which onig-config)
CGO_CFLAGS=$(shell $(ONIG_CONFIG) --cflags)
CGO_LDFLAGS=$(shell $(ONIG_CONFIG) --libs)

include $(GOROOT)/src/Make.pkg

prereq:
	@test -x "$(ONIG_CONFIG)" || (echo "Can't find onig_config in your path."; false)
