package main

import (
	"fmt"
	"log"
	"io"
        "strconv"
	"net/http"
        "image/jpeg"
	"archive/zip"
)

func main() {
  http.HandleFunc("/three-card/", threeHandler)
  log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func getRatio(img io.Reader) (float64, error) {
   dec, err := jpeg.Decode(img)
   if err != nil {
	return 0.0, err
   }
   b := dec.Bounds()
   mn, mx := b.Min, b.Max
   return (float64(mx.X - mn.X)/float64(mx.Y - mn.Y)), nil
} 

func getOrElse(lst []string, def string) string {
   if len(lst) > 0 {
       def = lst[0]  
   }
   return def
}

func threeHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "URL.PATH = %q\n", r.URL.Path)
  if err := r.ParseForm(); err != nil {
     log.Print(err)
  }
  
  desiredWidth, _ := strconv.Atoi(getOrElse(r.Form["w"],"600"))
  fmt.Fprintf(w, "Width = %v\n", desiredWidth)

  desiredDeck  := getOrElse(r.Form["deck"],"Lenormand")
  fmt.Fprintf(w, "Deck = %v\n", desiredDeck) 

  deck, err := zip.OpenReader(desiredDeck + ".zip")
  if err != nil {
     log.Print(err)
     return
  }
  defer deck.Close()
  
  fmt.Fprintf(w, "There are %d cards in the deck.\n", len(deck.File))

  // grab the first file and establish the aspect ratio of a card
  // (we will assume all cards are roughly the same shape)
  any, err := deck.File[0].Open()
  if err != nil {
     log.Print(err)
     return
  }
  rat, err := getRatio(any) 
  if err != nil {
     log.Print(err)
     return
  }
  any.Close()
  fmt.Fprintf(w, "Ratio of a file is %v\n",rat)
}

