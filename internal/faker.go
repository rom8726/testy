package internal

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FakerRegistry holds all faker functions
type FakerRegistry struct {
	functions map[string]FakerFunc
}

// FakerFunc is a function that generates fake data
type FakerFunc func() string

// NewFakerRegistry creates a new faker registry with default functions
func NewFakerRegistry() *FakerRegistry {
	registry := &FakerRegistry{
		functions: make(map[string]FakerFunc),
	}

	// Register default functions
	registry.RegisterDefaults()

	return registry
}

// RegisterDefaults registers all default faker functions
func (r *FakerRegistry) RegisterDefaults() {
	// UUID
	r.Register("uuid", GenerateUUID)
	r.Register("uuid.v4", GenerateUUID)

	// Names
	r.Register("name", GenerateName)
	r.Register("firstName", GenerateFirstName)
	r.Register("lastName", GenerateLastName)
	r.Register("fullName", GenerateName)

	// Internet
	r.Register("email", GenerateEmail)
	r.Register("username", GenerateUsername)
	r.Register("domain", GenerateDomain)
	r.Register("url", GenerateURL)
	r.Register("ipv4", GenerateIPv4)

	// Phone
	r.Register("phone", GeneratePhone)
	r.Register("phoneNumber", GeneratePhone)

	// Address
	r.Register("city", GenerateCity)
	r.Register("street", GenerateStreet)
	r.Register("country", GenerateCountry)
	r.Register("zipCode", GenerateZipCode)

	// Date/Time
	r.Register("date", GenerateDate)
	r.Register("time", GenerateTime)
	r.Register("timestamp", GenerateTimestamp)
	r.Register("now", GenerateNow)

	// Numbers
	r.Register("number", GenerateNumber)
	r.Register("integer", GenerateInteger)
	r.Register("float", GenerateFloat)
	r.Register("digit", GenerateDigit)

	// Text
	r.Register("word", GenerateWord)
	r.Register("words", GenerateWords)
	r.Register("sentence", GenerateSentence)
	r.Register("paragraph", GenerateParagraph)

	// Company
	r.Register("company", GenerateCompany)
	r.Register("companyName", GenerateCompany)

	// Random
	r.Register("random.string", GenerateRandomString)
	r.Register("random.int", GenerateRandomInt)
	r.Register("random.bool", GenerateRandomBool)
}

// Register adds a custom faker function
func (r *FakerRegistry) Register(name string, fn FakerFunc) {
	r.functions[name] = fn
}

// Generate generates fake data using the specified function name
func (r *FakerRegistry) Generate(name string) (string, error) {
	fn, exists := r.functions[name]
	if !exists {
		return "", fmt.Errorf("faker function not found: %s", name)
	}

	return fn(), nil
}

// === Generator Functions ===

// GenerateUUID generates a random UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateName generates a random full name
func GenerateName() string {
	return GenerateFirstName() + " " + GenerateLastName()
}

var firstNames = []string{
	"John", "Jane", "Michael", "Sarah", "David", "Emily", "Robert", "Lisa",
	"William", "Jessica", "James", "Ashley", "Daniel", "Amanda", "Joseph", "Melissa",
	"Thomas", "Deborah", "Charles", "Stephanie", "Christopher", "Rebecca", "Matthew", "Laura",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
	"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
	"Taylor", "Moore", "Jackson", "Martin", "Lee", "Thompson", "White", "Harris",
}

// GenerateFirstName generates a random first name
func GenerateFirstName() string {
	return randomChoice(firstNames)
}

// GenerateLastName generates a random last name
func GenerateLastName() string {
	return randomChoice(lastNames)
}

// GenerateEmail generates a random email address
func GenerateEmail() string {
	first := strings.ToLower(GenerateFirstName())
	last := strings.ToLower(GenerateLastName())
	domain := GenerateDomain()
	return fmt.Sprintf("%s.%s@%s", first, last, domain)
}

// GenerateUsername generates a random username
func GenerateUsername() string {
	first := strings.ToLower(GenerateFirstName())
	number := randomInt(100, 999)
	return fmt.Sprintf("%s%d", first, number)
}

var domains = []string{"example.com", "test.com", "mail.com", "email.com", "demo.com"}

// GenerateDomain generates a random domain
func GenerateDomain() string {
	return randomChoice(domains)
}

// GenerateURL generates a random URL
func GenerateURL() string {
	return "https://" + GenerateDomain()
}

// GenerateIPv4 generates a random IPv4 address
func GenerateIPv4() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		randomInt(1, 255),
		randomInt(0, 255),
		randomInt(0, 255),
		randomInt(0, 255))
}

// GeneratePhone generates a random phone number
func GeneratePhone() string {
	return fmt.Sprintf("+1-%03d-%03d-%04d",
		randomInt(100, 999),
		randomInt(100, 999),
		randomInt(1000, 9999))
}

var cities = []string{
	"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia",
	"San Antonio", "San Diego", "Dallas", "San Jose", "Austin", "Jacksonville",
}

// GenerateCity generates a random city name
func GenerateCity() string {
	return randomChoice(cities)
}

var streets = []string{
	"Main St", "Oak Ave", "Park Rd", "Elm St", "Washington St", "Maple Ave",
	"Lincoln Ave", "Cedar St", "Pine St", "Broadway", "Church St", "Hill Rd",
}

// GenerateStreet generates a random street name
func GenerateStreet() string {
	number := randomInt(1, 9999)
	street := randomChoice(streets)
	return fmt.Sprintf("%d %s", number, street)
}

var countries = []string{
	"United States", "Canada", "United Kingdom", "Germany", "France", "Italy",
	"Spain", "Australia", "Japan", "China", "India", "Brazil",
}

// GenerateCountry generates a random country name
func GenerateCountry() string {
	return randomChoice(countries)
}

// GenerateZipCode generates a random ZIP code
func GenerateZipCode() string {
	return fmt.Sprintf("%05d", randomInt(10000, 99999))
}

// GenerateDate generates a random date
func GenerateDate() string {
	days := randomInt(0, 365)
	date := time.Now().AddDate(0, 0, -days)
	return date.Format("2006-01-02")
}

// GenerateTime generates a random time
func GenerateTime() string {
	hour := randomInt(0, 23)
	minute := randomInt(0, 59)
	second := randomInt(0, 59)
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// GenerateTimestamp generates a random timestamp
func GenerateTimestamp() string {
	days := randomInt(0, 365)
	timestamp := time.Now().AddDate(0, 0, -days)
	return timestamp.Format(time.RFC3339)
}

// GenerateNow generates the current timestamp
func GenerateNow() string {
	return time.Now().Format(time.RFC3339)
}

// GenerateNumber generates a random number string
func GenerateNumber() string {
	return fmt.Sprintf("%d", randomInt(1, 1000000))
}

// GenerateInteger generates a random integer string
func GenerateInteger() string {
	return GenerateNumber()
}

// GenerateFloat generates a random float string
func GenerateFloat() string {
	integer := randomInt(1, 10000)
	decimal := randomInt(0, 99)
	return fmt.Sprintf("%d.%02d", integer, decimal)
}

// GenerateDigit generates a random single digit
func GenerateDigit() string {
	return fmt.Sprintf("%d", randomInt(0, 9))
}

var words = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
}

// GenerateWord generates a random word
func GenerateWord() string {
	return randomChoice(words)
}

// GenerateWords generates multiple random words
func GenerateWords() string {
	count := randomInt(3, 8)
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = GenerateWord()
	}
	return strings.Join(result, " ")
}

// GenerateSentence generates a random sentence
func GenerateSentence() string {
	sentence := GenerateWords()
	return strings.ToUpper(string(sentence[0])) + sentence[1:] + "."
}

// GenerateParagraph generates a random paragraph
func GenerateParagraph() string {
	sentences := randomInt(3, 6)
	result := make([]string, sentences)
	for i := 0; i < sentences; i++ {
		result[i] = GenerateSentence()
	}
	return strings.Join(result, " ")
}

var companies = []string{
	"Tech Corp", "Data Systems", "Cloud Solutions", "Digital Industries", "Innovation Labs",
	"Software Group", "Network Services", "Web Technologies", "Smart Systems", "Future Enterprises",
}

// GenerateCompany generates a random company name
func GenerateCompany() string {
	return randomChoice(companies)
}

// GenerateRandomString generates a random alphanumeric string
func GenerateRandomString() string {
	length := randomInt(8, 16)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

// GenerateRandomInt generates a random integer string
func GenerateRandomInt() string {
	return fmt.Sprintf("%d", randomInt(1, 1000000))
}

// GenerateRandomBool generates a random boolean string
func GenerateRandomBool() string {
	if randomInt(0, 1) == 0 {
		return "false"
	}
	return "true"
}

// === Helper Functions ===

func randomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(choices))))
	return choices[n.Int64()]
}

func randomInt(min, max int) int {
	if min >= max {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

// ExpandFakerPlaceholders expands faker placeholders in a string
// Format: {{faker.functionName}}
func ExpandFakerPlaceholders(input string, registry *FakerRegistry) string {
	// Simple replacement - in real implementation would use regex
	// For now, this is a placeholder for the concept
	return input
}
