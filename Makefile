include $(GOROOT)/src/Make.inc

ONIG_CONFIG=$(shell which onig-config)

prereq:
	@test -x "$(ONIG_CONFIG)" || (echo "Can't find onig-config in your path."; false)

oniguruma/Makefile:
	git submodule update --init

onig_install: oniguruma/Makefile
	cd oniguruma; ./configure --prefix=/usr/local; make install

rubex_install: lib/Makefile
	cd lib; make install

install: prereq onig_install rubex_install

test:
	cd lib; make test

clean:
	cd lib; make clean
	cd oniguruma; make clean
