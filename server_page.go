package main

import (
	"bytes"
	"io"
	"strconv"
	"text/template"
	"time"
)

type Header struct {
	Date   string
	Length int
}

type InfoPage struct {
	Used    string
	Count   string
	DayUsed string
	Percent string
}

type ErrorPage struct {
	Username string
}

// 文件大小单位转换，暂时未知文件大小类型
func HumanSize(size int64) (result string) {
	var n int64
	if size > 1073741824 {
		n = size / 1073741824
		result = strconv.FormatInt(n, 10) + "GB"
	} else if size > 1048576 {
		n = size / 1048576
		result = strconv.FormatInt(n, 10) + "MB"
	} else if size > 1024 {
		n = size / 1024
		result = strconv.FormatInt(n, 10) + "KB"
	} else {
		result = strconv.FormatInt(size, 10) + "Byte"
	}
	return
}

func DayOfMonth(year int, month time.Month) int {
	if month == time.April ||
		month == time.June ||
		month == time.September ||
		month == time.November {
		return 30
	}
	if month == time.February {
		if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
			return 29
		} else {
			return 28
		}
	}
	return 31
}

func WriteString(out io.Writer, str string) (n int, err error) {
	length := len(str)
	tmpl_header := "HTTP/1.1 200 OK\r\nServer: nginx/1.1.19\r\nDate: {{.Date}}\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: {{.Length}}\r\nLast-Modified: {{.Date}}\r\nConnection: keep-alive\r\nAccept-Ranges: bytes\r\n\r\n"
	header := &Header{time.Now().Format(time.RFC1123), length}
	tmpl, err := template.New("HtmlHeader").Parse(tmpl_header)
	if err != nil {
		return 0, nil
	}
	err = tmpl.Execute(out, header)
	if err != nil {
		return 0, nil
	}
	return out.Write([]byte(str))
}

func usedTraffic(used, count int64) map[string]interface{} {
	ctx := make(map[string]interface{})
	now := time.Now()
	left := count - used
	days := DayOfMonth(now.Year(), now.Month())
	// calc DayUsed
	dayused := int64(left) / int64(days)
	// calc Percent
	perval := float64(used) / float64(count) * 100
	if perval < 1 {
		perval = 1
	}
	percent := strconv.FormatFloat(perval, 'f', 0, 64)
	ctx["used"] = HumanSize(used)
	ctx["count"] = HumanSize(count)
	ctx["left"] = HumanSize(left)
	ctx["dayused"] = HumanSize(dayused)
	ctx["percent"] = percent
	return ctx
}

func WriteInfoPage(out io.Writer, used, count int64) (n int, err error) {
	var b bytes.Buffer
	now := time.Now()
	left := count - used
	days := DayOfMonth(now.Year(), now.Month())
	// calc DayUsed
	dayused := int64(left) / int64(days)
	// calc Percent
	perval := float64(used) / float64(count) * 100
	if perval < 1 {
		perval = 1
	}
	percent := strconv.FormatFloat(perval, 'f', 0, 64)
	info := &InfoPage{Used: HumanSize(used), Count: HumanSize(left), DayUsed: HumanSize(dayused), Percent: percent}
	tmpl, err := template.ParseFiles("templates/server/info.html")
	err = tmpl.Execute(&b, info)
	if err != nil {
		return 0, nil
	}
	return WriteString(out, b.String())
}

func WriteErrorPage(out io.Writer, user string) (n int, err error) {
	var b bytes.Buffer

	tmpl, err := template.ParseFiles("templates/server/error.html")
	header := &ErrorPage{Username: user}
	err = tmpl.Execute(&b, header)
	if err != nil {
		return 0, nil
	}
	return WriteString(out, b.String())
}
