package main

import (
	"fmt"
	"log"
        "strconv"
	"net/http"
        "image"
        "image/draw"
        "image/jpeg"
	"archive/zip"
        "github.com/nfnt/resize"
)

func main() {
  http.HandleFunc("/three-card/", threeHandler)
  log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func getRatio(imgs []*zip.File) (float64, error) {
   if len(imgs) < 1 {
      return 0.0, fmt.Errorf("No images in zip file!")
   }
   img, err := imgs[0].Open()
   if err != nil {
	return 0.0, err
   }
   defer img.Close()

   dec, err := jpeg.Decode(img)
   if err != nil {
	return 0.0, err
   }
   b := dec.Bounds()
   return (float64(b.Dx())/float64(b.Dy())), nil
} 

func getOrElse(lst []string, def string) string {
   if len(lst) > 0 {
       def = lst[0]  
   }
   return def
}

func threeHandler(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
     log.Print(err)
  }
  
  desiredWidth, _ := strconv.Atoi(getOrElse(r.Form["w"],"600"))
  desiredDeck  := getOrElse(r.Form["deck"],"Lenormand")
  deck, err := zip.OpenReader(desiredDeck + ".zip")
  if err != nil {
     log.Print(err)
     return
  }
  defer deck.Close()
  
  // grab the first file and establish the aspect ratio of a card
  // (we will assume all cards are roughly the same shape)
  rat, err := getRatio(deck.File) 
  if err != nil {
     log.Print(err)
     return
  }

  // numCards  := len(deck.File)
  cardWidth := (desiredWidth / 3) 
  cardHeight := int( float64(cardWidth)/rat )
  selected := []int{0,1,2}  // obviously not permanent

  answer := image.NewRGBA(image.Rect(0,0,cardWidth*len(selected),cardHeight))
  for idx, c := range selected {
    xloc := idx * cardWidth
    cardRect := image.Rect(xloc,0, xloc+cardWidth, cardHeight)

    // open the card...
    cardFile, err := deck.File[c].Open()
    if err != nil {
       log.Print(err)
       return
    }

    // read the JPG inside
    cardImg, err := jpeg.Decode(cardFile) 
    cardFile.Close()
    if err != nil {
       log.Print(err)
       return
    }

    // resize it...
    cardImg = resize.Resize(uint(cardWidth),uint(cardHeight),cardImg,resize.Bicubic)
    
    draw.Draw(answer, cardRect, cardImg, image.Pt(0,0), draw.Src)
  }
  err = jpeg.Encode(w, answer, &jpeg.Options{Quality: 80})
  if err != nil {
     log.Print(err)
  }
}

