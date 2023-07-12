package utilities

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func TS() {
	bs, _ := ioutil.ReadFile("d:\\temp\\2023\\05-15\\data.txt")
	s := string(bs)
	ss := strings.Split(s, "\r\n")
	res := ""
	for _, line := range ss {
		line = strings.Trim(line, "\r\n\t")
		if !strings.HasPrefix(line, "static const") {
			continue
		}
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == ' '
		})
		fmt.Println(parts)
		res += "icons[\"" + parts[3] + "\"] = Icons." + parts[3] + ";\r\n"
		// static const IconData zoom_in_map = IconData(0xf05af, fontFamily: 'MaterialIcons');

	}

	ioutil.WriteFile("d:\\temp\\2023\\05-15\\res.txt", []byte(res), 0777)
}
