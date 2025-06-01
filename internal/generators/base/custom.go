package base

import (
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// CustomPassword generates a random password with specified length
func CustomPassword(params map[string]interface{}) string {
	length := 12 // default length
	if val, ok := params["length"].(int); ok {
		length = val
	}

	// Define character sets
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	// Combine all character sets
	allChars := lowercase + uppercase + numbers + special

	// Ensure at least one character from each set
	password := []byte{
		lowercase[rand.Intn(len(lowercase))],
		uppercase[rand.Intn(len(uppercase))],
		numbers[rand.Intn(len(numbers))],
		special[rand.Intn(len(special))],
	}

	// Fill the rest randomly
	for i := 4; i < length; i++ {
		password = append(password, allChars[rand.Intn(len(allChars))])
	}

	// Shuffle the password
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password)
}

// CustomSentence generates a random sentence with specified word count
func CustomSentence(params map[string]interface{}) string {
	wordCount := 6 // default word count
	if val, ok := params["word_count"].(int); ok {
		wordCount = val
	}

	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = gofakeit.Word()
	}

	// Capitalize first word and add period
	sentence := strings.Join(words, " ")
	return strings.ToUpper(sentence[:1]) + sentence[1:] + "."
}

// CustomParagraph generates a random paragraph with specified sentence count
func CustomParagraph(params map[string]interface{}) string {
	sentenceCount := 3 // default sentence count
	if val, ok := params["sentence_count"].(int); ok {
		sentenceCount = val
	}

	sentences := make([]string, sentenceCount)
	for i := 0; i < sentenceCount; i++ {
		sentences[i] = CustomSentence(nil)
	}

	return strings.Join(sentences, " ")
}

// CustomDate generates a random date within a specified range
func CustomDate(params map[string]interface{}) time.Time {
	start := time.Now().AddDate(-1, 0, 0) // default: 1 year ago
	end := time.Now()                     // default: now

	if val, ok := params["start"].(time.Time); ok {
		start = val
	}
	if val, ok := params["end"].(time.Time); ok {
		end = val
	}

	// Calculate the duration between start and end
	duration := end.Sub(start)

	// Generate a random duration within the range
	randomDuration := time.Duration(rand.Int63n(int64(duration)))

	return start.Add(randomDuration)
}

// CustomNumber generates a random number within a specified range
func CustomNumber(params map[string]interface{}) int {
	min := 0   // default minimum
	max := 100 // default maximum

	if val, ok := params["min"].(int); ok {
		min = val
	}
	if val, ok := params["max"].(int); ok {
		max = val
	}

	return min + rand.Intn(max-min+1)
}

// CustomFloat generates a random float within a specified range
func CustomFloat(params map[string]interface{}) float64 {
	min := 0.0     // default minimum
	max := 100.0   // default maximum
	precision := 2 // default decimal places

	if val, ok := params["min"].(float64); ok {
		min = val
	}
	if val, ok := params["max"].(float64); ok {
		max = val
	}
	if val, ok := params["precision"].(int); ok {
		precision = val
	}

	// Generate random float
	value := min + rand.Float64()*(max-min)

	// Round to specified precision
	multiplier := float64(10 ^ precision)
	return float64(int64(value*multiplier)) / multiplier
}

// CustomString generates a random string with specified length and pattern
func CustomString(params map[string]interface{}) string {
	length := 10        // default length
	pattern := "??????" // default pattern

	if val, ok := params["length"].(int); ok {
		length = val
	}
	if val, ok := params["pattern"].(string); ok {
		pattern = val
	}

	// If pattern is provided, use it
	if pattern != "??????" {
		return gofakeit.Regex(pattern)
	}

	// Otherwise generate random string of specified length
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// CustomEmail generates a random email with specified domain
func CustomEmail(params map[string]interface{}) string {
	domain := "example.com" // default domain
	if val, ok := params["domain"].(string); ok {
		domain = val
	}

	username := gofakeit.Username()
	return username + "@" + domain
}

// CustomPhone generates a random phone number with specified format
func CustomPhone(params map[string]interface{}) string {
	format := "(###) ###-####" // default format
	if val, ok := params["format"].(string); ok {
		format = val
	}

	return gofakeit.Numerify(format)
}

// CustomCreditCard generates a random credit card with specified type
func CustomCreditCard(params map[string]interface{}) string {
	// Use gofakeit's CreditCardNumber directly
	return gofakeit.CreditCardNumber(nil)
}

// // parseInt tries to parse a string as int
// func parseInt(s string) (int, error) {
// 	var n int
// 	_, err := fmt.Sscanf(s, "%d", &n)
// 	return n, err
// }

// CustomShuffleStrings shuffles a list of strings
func CustomShuffleStrings(params map[string]interface{}) []string {
	// Get the list of strings to shuffle
	var stringList []string
	if val, ok := params["strings"].([]string); ok {
		stringList = val
	} else if val, ok := params["strings"].(string); ok {
		// If a single string is provided, split it by comma
		stringList = strings.Split(val, ",")
	}

	// Shuffle the strings
	rand.Shuffle(len(stringList), func(i, j int) {
		stringList[i], stringList[j] = stringList[j], stringList[i]
	})

	return stringList
}
