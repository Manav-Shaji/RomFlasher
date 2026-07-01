package platform

import "strings"

var deviceNameMap = map[string]string{
	"M2007J17I":  "MI 10I",
	"M2102K1G":   "MI 11 ULTRA",
	"M2102J20SG": "POCO X3 PRO",
	"M2012K11AG": "POCO F3 / MI 11X",
	"M2012K11AC": "REDMI K40 / POCO F3",
	"24069PC21I": "POCO F6",
	"24069PC21G": "POCO F6",
	"PERIDOT":    "POCO F6",
	"BERYLLIUM":  "POCO F1",
	"MONDRIAN":   "POCO F5 PRO",
	"MARBLE":     "POCO F5",
	"VILI":       "XIAOMI 11T PRO",
	"SWEET":      "REDMI NOTE 10 PRO",
	"SUNNY":      "REDMI NOTE 10",
}

func prettyDeviceName(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	if name, ok := deviceNameMap[id]; ok {
		return name
	}
	return id
}
