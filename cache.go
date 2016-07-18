package main

// a simple cache so we don't keep reloading the same deck

import (
	"sync"
)

const (
	safeDeck = "Poker.zip" // part of the distribution
)

type cacheEnt struct {
	deck   *deck
	refcnt uint32
}

var cacheLock sync.Mutex
var cache map[string]*cacheEnt

func requestDeck(name string) (answer *deck, err error) {
	cacheLock.Lock()

	if cache == nil {
		cache = make(map[string]*cacheEnt)
	}

	ent, ok := cache[name]
	if ok {
		answer = ent.deck
		ent.refcnt++
	} else {
		answer, err = NewDeck(name)
		if err == nil {
			ent = &cacheEnt{deck: answer, refcnt: 1}
			cache[name] = ent
		}
	}
	cacheLock.Unlock()

	// try to fall back on the poker deck... it's safe!
	if err != nil && name != safeDeck {
		return requestDeck(safeDeck)
	}

	return
}

func returnDeck(d *deck) {
	cacheLock.Lock()

	if ent, ok := cache[d.name]; ok {
		if ent.refcnt <= 1 {
			d.Close()
			delete(cache, d.name)
		}
		ent.refcnt--
	} else {
		// we couldn't find the deck, so
		// better close it to be on the safe side
		d.Close()
	}

	cacheLock.Unlock()
}
