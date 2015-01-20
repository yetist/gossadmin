package main

import (
	"github.com/chai2010/gettext-go/gettext"
	"strings"
)

var (
	sep = '\004'
)

func SetLocale(locale string) string {
	return gettext.SetLocale(locale)
}

func BindTextdomain(domain string, path string, zipData []byte) ([]string, []string) {
	return gettext.BindTextdomain(domain, path, zipData)
}

func Textdomain(domain string) string {
	return gettext.Textdomain(domain)
}

// the string to be translated, with a '|'-separated prefix
// which must not be translated
func Q_(msgid string) string {
	return pGettext(msgid, 0)
}

func __(msgid string) string {
	return gettext.PGettext("", msgid)
}

// Only marks a string for translation. This is useful in situations
// where the translated strings can't be directly used, e.g. in string
// array initializers. To get the translated string, call gettext()
// at runtime.
func N_(msgid string) string {
	return msgid
}

// Only marks a string for translation with context.
func NC_(context, msgid string) string {
	return pGettext2(context, msgid)
}

// A context is a prefix to your translation, usefull when one word has different meanings, depending on its context.
// C_("Printer","Open") <=> C_("File","Open")
// is the same as Q_("Printer|Open")  <=> Q_("File|Open")
func C_(context, msgid string) string {
	return gettext.PGettext(context, msgid)
}

func D_(name string) []byte {
	return gettext.Getdata(name)
}

func pGettext(msgctxtid string, msgidoffset uint64) string {
	translation := gettext.Gettext(msgctxtid)
	if translation == msgctxtid {
		if msgidoffset > 0 {
			return msgctxtid[msgidoffset:]
		}
		var pos int
		if pos = strings.Index(msgctxtid, "|"); pos > -1 {
			tmp := strings.Replace(msgctxtid, "|", string(sep), 1)
			translation = gettext.Gettext(tmp)
			if translation == tmp {
				return translation[pos+1:]
			}
		}
	}
	return translation
}

func pGettext2(msgctxt, msgid string) string {
	msg_ctxt_id := msgctxt + string(sep) + msgid
	translation := gettext.Gettext(msg_ctxt_id)

	if translation == msg_ctxt_id {
		tmp := strings.Replace(msg_ctxt_id, string(sep), "|", 1)
		translation = gettext.Gettext(tmp)

		if translation == tmp {
			return msgid
		}
	}
	return translation
}
