#!/bin/bash
olddir=$PWD
if [ ! -f gossadmin ]; then
	godep go build
fi
if [ -f gossadmin ]; then
	./gossadmin -config 2>&1 > /dev/null
	TMP=`mktemp -d`
	mkdir -p ${TMP}/usr/{share/{gossadmin,locale/zh_CN/LC_MESSAGES},bin}
	mkdir -p ${TMP}/etc/xdg/gossadmin/0.1/
	install -m755 gossadmin ${TMP}/usr/bin/gossadmin
	cp -r assets ${TMP}/usr/share/gossadmin/
	cp -r templates ${TMP}/usr/share/gossadmin/
	cp locale/zh_CN/LC_MESSAGES/gossadmin.mo ${TMP}/usr/share/locale/zh_CN/LC_MESSAGES/
	cp server.ini ${TMP}/etc/xdg/gossadmin/0.1/
	rm -f server.ini
	cd ${TMP}
	tar -zcf $olddir/gossadmin.tar.gz *
	echo "gossadmin.tar.gz" is ready
fi
