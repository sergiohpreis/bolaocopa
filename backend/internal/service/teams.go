package service

var teamNames = map[string]string{
	"Algeria":            "Argélia",
	"Argentina":          "Argentina",
	"Australia":          "Austrália",
	"Austria":            "Áustria",
	"Belgium":            "Bélgica",
	"Bosnia-Herzegovina": "Bósnia-Herzegovina",
	"Brazil":             "Brasil",
	"Canada":             "Canadá",
	"Cape Verde Islands": "Cabo Verde",
	"Colombia":           "Colômbia",
	"Congo DR":           "Congo",
	"Croatia":            "Croácia",
	"Curaçao":            "Curaçao",
	"Czechia":            "República Tcheca",
	"Ecuador":            "Equador",
	"Egypt":              "Egito",
	"England":            "Inglaterra",
	"France":             "França",
	"Germany":            "Alemanha",
	"Ghana":              "Gana",
	"Haiti":              "Haiti",
	"Iran":               "Irã",
	"Iraq":               "Iraque",
	"Ivory Coast":        "Costa do Marfim",
	"Japan":              "Japão",
	"Jordan":             "Jordânia",
	"Mexico":             "México",
	"Morocco":            "Marrocos",
	"Netherlands":        "Holanda",
	"New Zealand":        "Nova Zelândia",
	"Norway":             "Noruega",
	"Panama":             "Panamá",
	"Paraguay":           "Paraguai",
	"Portugal":           "Portugal",
	"Qatar":              "Catar",
	"Saudi Arabia":       "Arábia Saudita",
	"Scotland":           "Escócia",
	"Senegal":            "Senegal",
	"South Africa":       "África do Sul",
	"South Korea":        "Coreia do Sul",
	"Spain":              "Espanha",
	"Sweden":             "Suécia",
	"Switzerland":        "Suíça",
	"Tunisia":            "Tunísia",
	"Turkey":             "Turquia",
	"United States":      "Estados Unidos",
	"Uruguay":            "Uruguai",
	"Uzbekistan":         "Uzbequistão",
}

func translateTeam(name string) string {
	if pt, ok := teamNames[name]; ok {
		return pt
	}
	return name
}
