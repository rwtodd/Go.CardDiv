package main

// a simple cache so we don't keep reloading the same deck

import "sync"

const (
	safeDeck = "Poker.zip" // part of the distribution
)

var cacheLock sync.Mutex
var latest *deck

func requestDeck(name string) (*deck, error) {
	var (
		answer *deck
		err    error
	)

	fullname, err := rscBase.Path(name)
	if err != nil {
		return nil, err
	}

	cacheLock.Lock()

	// is it already our latest deck?  If so, return it.
	if latest != nil {
		if latest.name == fullname {
			// the cached deck matches; return it
			answer = latest
		} else {
			// it didn't match, clear it out
			latest.Close()
			latest = nil
		}
	}

	if answer == nil {
		// it wasn't in the cache, so look up the deck
		latest, err = newDeck(fullname)
		if err == nil {
			latest.Open()
			answer = latest
		}
	}

	// if we have an answer, Open() it on behalf of the caller
	// so that it has an immediate reference
	if answer != nil {
		answer.Open()
	}

	cacheLock.Unlock()

	if answer == nil && name != safeDeck {
		// we couldn't find the requested deck,
		// so fall back on the one that should always
		// be there
		return requestDeck(safeDeck)
	}

	return answer, err
}
