package timezone

import (
	"log"
	"time"
)

var countryToTZ = map[string]*time.Location{}

func Load() {
	loadLocation := func(tz string) *time.Location {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Fatalf("Failed to load location %s: %v", tz, err)
		}
		return loc
	}

	countryToTZ = map[string]*time.Location{
		"Albania":                loadLocation("Europe/Tirane"),
		"Andorra":                loadLocation("Europe/Andorra"),
		"Austria":                loadLocation("Europe/Vienna"),
		"Belarus":                loadLocation("Europe/Minsk"),
		"Belgium":                loadLocation("Europe/Brussels"),
		"Bosnia and Herzegovina": loadLocation("Europe/Sarajevo"),
		"Bulgaria":               loadLocation("Europe/Sofia"),
		"Croatia":                loadLocation("Europe/Zagreb"),
		"Cyprus":                 loadLocation("Asia/Nicosia"),
		"Czech Republic":         loadLocation("Europe/Prague"),
		"Denmark":                loadLocation("Europe/Copenhagen"),
		"Estonia":                loadLocation("Europe/Tallinn"),
		"Finland":                loadLocation("Europe/Helsinki"),
		"France":                 loadLocation("Europe/Paris"),
		"Germany":                loadLocation("Europe/Berlin"),
		"Greece":                 loadLocation("Europe/Athens"),
		"Hungary":                loadLocation("Europe/Budapest"),
		"Iceland":                loadLocation("Atlantic/Reykjavik"),
		"Ireland":                loadLocation("Europe/Dublin"),
		"Italy":                  loadLocation("Europe/Rome"),
		"Latvia":                 loadLocation("Europe/Riga"),
		"Liechtenstein":          loadLocation("Europe/Vaduz"),
		"Lithuania":              loadLocation("Europe/Vilnius"),
		"Luxembourg":             loadLocation("Europe/Luxembourg"),
		"Malta":                  loadLocation("Europe/Malta"),
		"Moldova":                loadLocation("Europe/Chisinau"),
		"Monaco":                 loadLocation("Europe/Monaco"),
		"Montenegro":             loadLocation("Europe/Podgorica"),
		"Netherlands":            loadLocation("Europe/Amsterdam"),
		"North Macedonia":        loadLocation("Europe/Skopje"),
		"Norway":                 loadLocation("Europe/Oslo"),
		"Poland":                 loadLocation("Europe/Warsaw"),
		"Portugal":               loadLocation("Europe/Lisbon"),
		"Romania":                loadLocation("Europe/Bucharest"),
		"Russia":                 loadLocation("Europe/Moscow"),
		"San Marino":             loadLocation("Europe/San_Marino"),
		"Serbia":                 loadLocation("Europe/Belgrade"),
		"Slovakia":               loadLocation("Europe/Bratislava"),
		"Slovenia":               loadLocation("Europe/Ljubljana"),
		"Spain":                  loadLocation("Europe/Madrid"),
		"Sweden":                 loadLocation("Europe/Stockholm"),
		"Switzerland":            loadLocation("Europe/Zurich"),
		"Turkey":                 loadLocation("Europe/Istanbul"),
		"Ukraine":                loadLocation("Europe/Kyiv"),
		"United Kingdom":         loadLocation("Europe/London"),
		"Vatican City":           loadLocation("Europe/Vatican"),
	}
}

func Transform(time time.Time, country string) (time.Time, bool) {
	location, ok := countryToTZ[country]
	if !ok {
		return time, false
	}
	return time.In(location), true
}
