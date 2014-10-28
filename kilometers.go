package km

import "time"

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

type Field struct {
	Km   int
	Time string
	Name string
}

func (k *Kilometers) getMax() int {
	if k.Terug > 0 {
		return k.Terug
	}
	if k.Laatste > 0 {
		return k.Laatste
	}
	if k.Eerste > 0 {
		return k.Eerste
	}
	if k.Begin > 0 {
		return k.Begin
	}
	return 0
}

func (k *Kilometers) AddFields(fields []Field) {
	for _, field := range fields {
		switch field.Name {
		case "Begin":
			k.Begin = field.Km
		case "Eerste":
			k.Eerste = field.Km
		case "Laatste":
			k.Laatste = field.Km
		case "Terug":
			k.Terug = field.Km
		}

	}
}
