package random

import (
	"math/rand"
	"time"
)

var (
	left = []string{
		"admiring",
		"adoring",
		"affectionate",
		"amazing",
		"awesome",
		"beautiful",
		"blissful",
		"bold",
		"brave",
		"charming",
		"clever",
		"compassionate",
		"competent",
		"confident",
		"cool",
		"dazzling",
		"determined",
		"dreamy",
		"eager",
		"ecstatic",
		"elastic",
		"elated",
		"elegant",
		"eloquent",
		"epic",
		"exciting",
		"fervent",
		"festive",
		"flamboyant",
		"focused",
		"friendly",
		"funny",
		"gallant",
		"gifted",
		"gracious",
		"great",
		"happy",
		"hardcore",
		"heuristic",
		"hopeful",
		"hungry",
		"infallible",
		"inspiring",
		"intelligent",
		"interesting",
		"jolly",
		"jovial",
		"keen",
		"kind",
		"laughing",
		"loving",
		"lucid",
		"magical",
		"modest",
		"musing",
		"mystifying",
		"nice",
		"nifty",
		"nostalgic",
		"objective",
		"optimistic",
		"peaceful",
		"pensive",
		"practical",
		"priceless",
		"quirky",
		"quizzical",
		"recursing",
		"relaxed",
		"reverent",
		"romantic",
		"serene",
		"sharp",
		"sleepy",
		"stoic",
		"sweet",
		"tender",
		"thirsty",
		"trusting",
		"unruffled",
		"upbeat",
		"vibrant",
		"vigilant",
		"vigorous",
		"wizardly",
		"wonderful",
		"xenodochial",
		"youthful",
		"zealous",
		"zen",
	}

	// feel free to add your first name or GitHub username here if you are a contributor to CasaOS :)
	right = []string{
		"allen",
		"andres",
		"angelina",
		"austin",
		"bobo",
		"et",
		"ezreal",
		"grace",
		"jerry",
		"john",
		"lauren",
		"link",
		"oscar",
		"rally",
		"tiger",
		"xuhai",
	}
)

// GetRandomName generates a random name from the list of adjectives and surnames in this package formatted as "adjective_surname".
func Name(suffix *string) string {
	name := left[rand.Intn(len(left))] + "_" + right[rand.Intn(len(right))] //nolint:gosec // G404: Use of weak random number generator (math/rand instead of crypto/rand)

	if suffix != nil {
		name += "_" + *suffix
	}

	return name
}

func String(n int, onlyLetters bool) string {
	return RandomString(n, onlyLetters)
}

// Deprecated: use random.String(...) instead
func RandomString(n int, onlyLetter bool) string { //nolint:revive
	var letters []rune

	if onlyLetter {
		letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	} else {
		letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	}

	b := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec // G404: Use of weak random number generator (math/rand instead of crypto/rand)
	}
	return string(b)
}
