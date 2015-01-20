#!/bin/bash
LANGS=('zh_CN')
DOMAIN=gossadmin
KEYWORDS="--keyword=__ --keyword=N_ --keyword=C_:1c,2 --keyword=NC_:1c,2 --keyword=Q_:1g"

update_sources() {
	for src in `find . -name "*.go"`;do
		xgettext -s -i -L C++ --from-code=utf-8 -d $DOMAIN -o ${DOMAIN}_src.pot $KEYWORDS $src
		if [ -f ${DOMAIN}_src.pot ];then
			[ -f ${DOMAIN}.pot ] || cp ${DOMAIN}_src.pot ${DOMAIN}.pot 
			msgcat ${DOMAIN}.pot ${DOMAIN}_src.pot -o ${DOMAIN}.pot
			rm -f ${DOMAIN}_src.pot
		fi
	done
}

update_tmpl() {
	tmpls=`find templates -name "*.html" -o -name "*.tmpl"`
	for tmpl in $tmpls; do
		xgettext -s -i -L Python --from-code=utf-8 -d $DOMAIN -o ${DOMAIN}_tmpl.pot $KEYWORDS $tmpl
		if [ -f ${DOMAIN}_tmpl.pot ];then
			[ -f ${DOMAIN}.pot ] || cp ${DOMAIN}_tmpl.pot ${DOMAIN}.pot 
			msgcat ${DOMAIN}.pot ${DOMAIN}_tmpl.pot -o ${DOMAIN}.pot
			rm -f ${DOMAIN}_tmpl.pot
		fi
	done
}

update_po() {
	for l in ${LANGS[@]}; do
		[ -d locale/$l/LC_MESSAGES ] || mkdir -p locale/$l/LC_MESSAGES
		if [ -f locale/$l/LC_MESSAGES/${DOMAIN}.po ]; then
			msgmerge -U locale/$l/LC_MESSAGES/${DOMAIN}.po ${DOMAIN}.pot
		else
			msginit -l ${l}.utf8 -o locale/$l/LC_MESSAGES/${DOMAIN}.po -i ${DOMAIN}.pot
		fi
		msgfmt -c -v -o  locale/$l/LC_MESSAGES/${DOMAIN}.mo locale/$l/LC_MESSAGES/${DOMAIN}.po
	done
}
[ -f ${DOMAIN}.pot ] && cp ${DOMAIN}.pot ${DOMAIN}.bak.pot 
update_sources
update_tmpl
update_po
