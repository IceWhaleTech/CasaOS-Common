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
		"breathtaking",
		"bright",
		"brilliant",
		"charismatic",
		"charming",
		"cheerful",
		"clever",
		"compassionate",
		"competent",
		"confident",
		"cool",
		"courageous",
		"dazzling",
		"delightful",
		"determined",
		"dreamy",
		"eager",
		"ecstatic",
		"effervescent",
		"elegant",
		"eloquent",
		"empathetic",
		"empowering",
		"encouraging",
		"energetic",
		"enlightened",
		"enthusiastic",
		"epic",
		"excellent",
		"exciting",
		"exquisite",
		"fascinating",
		"fearless",
		"fervent",
		"festive",
		"flamboyant",
		"flourishing",
		"focused",
		"friendly",
		"funny",
		"gallant",
		"generous",
		"gentle",
		"gifted",
		"glorious",
		"graceful",
		"gracious",
		"great",
		"happy",
		"hardcore",
		"harmonious",
		"heartwarming",
		"heuristic",
		"hopeful",
		"hungry",
		"illustrious",
		"imaginative",
		"impressive",
		"incredible",
		"infallible",
		"inspiring",
		"intelligent",
		"interesting",
		"intuitive",
		"invigorated",
		"invigorating",
		"joyful",
		"jovial",
		"jubilant",
		"keen",
		"kind",
		"laughing",
		"lively",
		"loving",
		"lucid",
		"magical",
		"majestic",
		"marvelous",
		"mesmerizing",
		"mindful",
		"modest",
		"musing",
		"mystifying",
		"noble",
		"nurturing",
		"optimistic",
		"passionate",
		"patient",
		"peaceful",
		"picturesque",
		"playful",
		"pleasant",
		"poetic",
		"positive",
		"powerful",
		"practical",
		"priceless",
		"proud",
		"pure",
		"quirky",
		"quizzical",
		"radiant",
		"ravishing",
		"recursing",
		"refreshing",
		"relaxed",
		"renewed",
		"resilient",
		"reverent",
		"romantic",
		"serene",
		"sharp",
		"shining",
		"sincere",
		"soothing",
		"sparkling",
		"spectacular",
		"splendid",
		"steadfast",
		"stirring",
		"stoic",
		"stunning",
		"sublime",
		"successful",
		"sweet",
		"tender",
		"thriving",
		"trusting",
		"unruffled",
		"upbeat",
		"uplifting",
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
