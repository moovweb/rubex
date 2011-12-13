include $(GOROOT)/src/Make.inc

ONIG_CONFIG=$(shell which onig-config)

prereq:
	@test -x "$(ONIG_CONFIG)" || (echo "Can't find onig-config in your path. do \"make onig_install\""; false)

oniguruma/Makefile:
	git submodule update --init

onig_install: oniguruma/Makefile
	cd oniguruma; ./configure --prefix=/usr/local; make; sudo make install
	echo "you may need to add \"export  LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:/usr/local/lib\" to your .bashrc or .profile"

rubex_install: lib/Makefile
	cd lib; make install

install: prereq rubex_install

test:
	cd lib; make test

clean:
	cd lib; make clean
	cd oniguruma; make clean
