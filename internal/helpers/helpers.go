package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/patrickoliveros/bookings/internal/config"
	"golang.org/x/crypto/bcrypt"
)

var AppConfig *config.AppConfig

type IntRange struct {
	min, max int
}

func HandleError(err error) {
	if err != nil {
		log.Println(err)
		return
	}
}

func HandleFatalError(err error, errorMessage string) error {
	if err != nil {
		log.Fatal(errorMessage)
		return err
	}

	return nil
}

func HandleDate(strInput string) (time.Time, string, error) {
	layout := "2006-01-02"
	layoutUS := "January 2, 2006"

	t, err := time.Parse(layout, strInput)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot parse date: %s", strInput))
		return t, "", err
	}

	readableDate := t.Format(layoutUS)

	return t, readableDate, nil
}

func SanitizeString(strInput string) string {

	strInput = strings.Replace(strInput, "-", " ", -1)

	return strings.Title(strInput)
}

func GenerateHashedPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	return string(hashedPassword)
}

func GenerateGuid() string {
	rawGuid := uuid.New()
	guid := strings.Replace(rawGuid.String(), "-", "", -1)

	generatedGuid := sanitizeGuid(guid)
	return strings.ToUpper(generatedGuid[:8])
}

func sanitizeGuid(s string) string {
	firstCharacter := rune(s[0])
	if !unicode.IsDigit(firstCharacter) {
		return fmt.Sprintf("%s%s", getAlphaCharacter(), s[1:])
	}

	return s
}

func LogObject(prefix string, obj interface{}) {
	jsonitem, _ := json.MarshalIndent(obj, " ", " ")

	log.Println(prefix, string(jsonitem))
}

func getAlphaCharacter() string {

	rn := rand.New(rand.NewSource(time.Now().UnixNano()))
	ir := IntRange{65, 90}

	genRan := ir.NextRandom(rn)

	return string(rune(genRan))
}

func (ir *IntRange) NextRandom(r *rand.Rand) int {
	return r.Intn(ir.max-ir.min+1) + ir.min
}

func GetCalendarDates(t time.Time) (string, string, string, string, string, string) {

	next := t.AddDate(0, 1, 0)
	last := t.AddDate(0, -1, 0)

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	thisMonth := t.Format("01")
	thisMonthYear := t.Format("2006")

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	return lastMonth, lastMonthYear, thisMonth, thisMonthYear,
		nextMonth, nextMonthYear

}

func GetMonthBoundaries(now time.Time) (time.Time, time.Time) {

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	return firstOfMonth, lastOfMonth

}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
