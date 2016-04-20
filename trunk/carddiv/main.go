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
	http.HandleFunc("/celtic/", celticHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func getOrElse(lst []string, def string) string {
	if len(lst) > 0 {
		def = lst[0]
	}
	return def
}

// rowHandler generates an image of cards in a row, with optional
// overlap.
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
		var co cardOpts
		if rand.Float64() >= 0.5 {
			co.reversed = true
		}
		cardImg, err := deck.Image(c, cardWidth, co)
		if err != nil {
			log.Print(err)
			return
		}

		draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)
	}
	err = jpeg.Encode(w, answer, &jpeg.Options{Quality: 80})
	if err != nil {
		log.Print(err)
	}
}

// celticHandler generates an image of cards in a celtic cross.
func celticHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}

	desiredWidth, _ := strconv.Atoi(getOrElse(r.Form["width"], "600"))
	desiredDeck := getOrElse(r.Form["deck"], "Lenormand")
	log.Printf("CELTIC:  Deck: %s Width: %d",
		desiredDeck,
		desiredWidth)

	deck, err := requestDeck(desiredDeck + ".zip")
	if err != nil {
		log.Print(err)
		return
	}

        // the overall image is 7 cards wide and 4 tall:
        //  0123456
        // |  x   x|
        // |x x x x|   the "cross" part is lowered by 
        // |  x   x|   half a card, relative to this pic.
        // |      x|
	cardWidth := uint(float64(desiredWidth) / 7.0)
        cardSize := image.Point{int(cardWidth), int(deck.CardHeight(cardWidth))} 

	// now, shuffle the deck
	selected, err := deck.Shuffled(10)
	if err != nil {
		log.Print(err)
		return
	}

	// now, create the image
	actualWidth := 7 * cardSize.X
        actualHeight := 4 * cardSize.Y
	answer := image.NewRGBA(image.Rect(0, 0, actualWidth, actualHeight))

        // nested helper function to create the card images
        getImage := func(which int, side bool) (image.Image, error) {
           var co cardOpts
	   if rand.Float64() >= 0.5 {
		co.reversed = true
	   }
           co.onSide = side
           return  deck.Image(selected[which], cardWidth, co)
        }

        // draw the cross...
        // 1. middle
        cardImg, err := getImage(0, false) 
        if err != nil {
		log.Print(err)
		return
        }
        midCard := image.Point{ (actualWidth - 2*cardSize.X)/2, 
                                (actualHeight - cardSize.Y)/2 }
        cardRect :=  image.Rectangle{midCard, midCard.Add(cardSize)}
	draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)

        // 2. crosses it
        cardImg, err = getImage(1, true) 
        if err != nil {
		log.Print(err)
		return
        }
        cardLoc := image.Point{ (actualWidth - cardSize.X - cardSize.Y)/2,
                                (actualHeight - cardSize.X)/2 }
        cardRect = image.Rectangle{cardLoc, 
                                   cardLoc.Add(image.Pt(cardSize.Y,cardSize.X))}
	draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)
         
        // 3. below it
        cardImg, err = getImage(2, false) 
        if err != nil {
		log.Print(err)
		return
        }
        cardLoc = midCard.Add(image.Pt(0,cardSize.Y + cardSize.Y/3)) 
        cardRect = image.Rectangle{cardLoc, cardLoc.Add(cardSize)} 
	draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)
        
        // 4. Waning Influence
        cardImg, err = getImage(3, false) 
        if err != nil {
		log.Print(err)
		return
        }
        cardLoc = midCard.Sub(image.Pt(cardSize.X*2,0)) 
        cardRect = image.Rectangle{cardLoc, cardLoc.Add(cardSize)} 
	draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)

        // 5. New Energy 
        cardImg, err = getImage(4, false) 
        if err != nil {
		log.Print(err)
		return
        }
        cardLoc = midCard.Sub(image.Pt(0,cardSize.Y + cardSize.Y/3)) 
        cardRect = image.Rectangle{cardLoc, cardLoc.Add(cardSize)} 
	draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)

        // 6. Waxing Influence
        cardImg, err = getImage(5, false) 
        if err != nil {
		log.Print(err)
		return
        }
        cardLoc = midCard.Add(image.Pt(cardSize.X*2,0)) 
        cardRect = image.Rectangle{cardLoc, cardLoc.Add(cardSize)} 
	draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)

        // 7 through 10...
        cardLoc = image.Point{cardSize.X*6,cardSize.Y*3}
        for idx := 6; idx < 10; idx++ {
           cardImg, err = getImage(idx, false) 
           if err != nil {
		log.Print(err)
		return
           }
           cardRect = image.Rectangle{cardLoc, cardLoc.Add(cardSize)} 
	   draw.Draw(answer, cardRect, cardImg, image.ZP, draw.Src)
           cardLoc = cardLoc.Sub(image.Pt(0,cardSize.Y)) 
        } 

	err = jpeg.Encode(w, answer, &jpeg.Options{Quality: 80})
	if err != nil {
		log.Print(err)
	}
}
