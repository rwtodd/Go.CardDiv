package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// hold a global cached deck between requests...
var curDeck *deck

func requestDeck(name string) (*deck, error) {
	if curDeck != nil && curDeck.Name() == name {
		return curDeck, nil
	}

	if curDeck != nil {
		curDeck.Close()
	}
	var err error
	curDeck, err = NewDeck(name)
	return curDeck, err
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/row/", rowHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func getOrElse(lst []string, def string) string {
	if len(lst) > 0 {
		def = lst[0]
	}
	return def
}

func rowHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}

	desiredWidth, _ := strconv.Atoi(getOrElse(r.Form["width"], "600"))
	desiredCards, _ := strconv.Atoi(getOrElse(r.Form["cards"], "3"))
	desiredShowing, _ := strconv.Atoi(getOrElse(r.Form["pct"], "100"))
	desiredDeck := getOrElse(r.Form["deck"], "Lenormand")
	log.Printf("ROW:  Deck: %s Cards: %d  Width: %d  Showing: %d%%",
		desiredDeck,
		desiredCards,
		desiredWidth,
		desiredShowing)

	deck, err := requestDeck(desiredDeck + ".zip")
	if err != nil {
		log.Print(err)
		return
	}

	// to account for overlap, we figure out the number of
	// cards effectively showing.  Thus 3 cards showing at 100%
	// would be 1 + 1 + 1, while at 80% it would be .8 + .8 + 1
	// (since the last card is fully visible)
	showPct := float64(desiredShowing) / 100.0
	effectiveCards := 1.0 + float64(desiredCards-1)*showPct
	cardWidth := uint(float64(desiredWidth) / effectiveCards)
	cardHeight := deck.CardHeight(cardWidth)
	showingWidth := int(float64(cardWidth) * showPct)

	// now, shuffle the deck
	selected, err := deck.Shuffled(desiredCards)
	if err != nil {
		log.Print(err)
		return
	}

	// now, create the image
	actualWidth := int(effectiveCards * float64(cardWidth))
	answer := image.NewRGBA(image.Rect(0, 0, actualWidth, int(cardHeight)))
	for idx, c := range selected {
		xloc := idx * showingWidth
		cardRect := image.Rect(xloc, 0, xloc+int(cardWidth), int(cardHeight))

		// open the card...
		cardImg, err := deck.Image(c, cardWidth, nil)
		if err != nil {
			log.Print(err)
			return
		}

		draw.Draw(answer, cardRect, cardImg, image.Pt(0, 0), draw.Src)
	}
	err = jpeg.Encode(w, answer, &jpeg.Options{Quality: 80})
	if err != nil {
		log.Print(err)
	}
}
