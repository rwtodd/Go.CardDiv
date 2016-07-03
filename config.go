package main

// configuration to send to the HTML/JS portion of the
// app, which keeps all the logic needed to add another
// layout in the Go part, unless new parameters need
// to be added.

type cardConfig struct {
	ID      string
	Display string
	Params  []string
}

var configurations = []cardConfig{
	{"/carddiv/row/", "Row of Cards", []string{"cards", "pct" }},
	{"/carddiv/celtic/", "Celtic Cross", []string{}},
	{"/carddiv/houses/", "12 Astrological Houses", []string{}},
	{"/carddiv/tableau/", "Grand Tableau", []string{}},
}
