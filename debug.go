package utilz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/maphash"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aybabtme/rgbterm"
	tm "github.com/buger/goterm"
	"github.com/davecgh/go-spew/spew"
	"github.com/hako/durafmt"
)

// TODO: number to color
// TODO: string to color
// TODO: short readable hash
const (
	Checkmark = "✓"
	XMark     = "✗"
)

func Sf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
func Ln(a ...interface{}) {
	fmt.Println(a...)
}

// Sfln is alias of fmt.Println(fmt.Sprintf())
func Sfln(format string, a ...interface{}) {
	Ln(Sf(format, a...))
}
func Errorln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stderr, a...)
}

var (
	DebugPrefix   string = "[DEBU]"
	InfoPrefix    string = "[INFO]"
	SuccessPrefix string = Lime("[SUCC]")
	WarnPrefix    string = Yellow("[WARN]")
	ErrorPrefix   string = RedBG("[ERRO]")
	FatalPrefix   string = RedBG("[FATAL]")
)
var (
	LogIncludeLevel bool = true
)

// GetCallerLocation returns the source location of the call at callDepth stack
// frames above the call.
func GetCallerLocation(callDepth int) (string, int) {
	_, file, line, _ := runtime.Caller(callDepth + 1)
	return getBaseFilename(file), line
}

func ThisFuncName() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	fn := runtime.FuncForPC(pc[0])
	return fn.Name()
}

func getBaseFilename(filename string) string {
	return filepath.Base(filename)
}

type LogHeaderParameter func() string

// LogParamCallStack adds the file and line number of the log call to the log message.
func LogParamCallStack() string {
	file, line := GetCallerLocation(4)
	return Sf("%s:%v", file, line)
}

// LogParamTimestamp adds a timestamp of current time to the log header.
func LogParamTimestamp() string {
	return KitchenTimeNow()
}

// LogParamTimestamp adds a timestamp of current time (that inludes milliseconds) to the log header.
func LogParamTimestampMs() string {
	return KitchenTimeMsNow()
}

// LogMessageNumber adds the index of the log call (incremental) to the log header.
func LogMessageNumber() string {
	return strconv.FormatInt(atomic.LoadInt64(&logMessageCounter), 10)
}

// LogElapsedFromStart adds the time passed from the program start.
func LogElapsedFromStart() string {
	return formatDuration(time.Now().Sub(logStartedAt))
}

// LogElapsedFromLastLogMessage adds to the log header the time passed from the last log call.
func LogElapsedFromLastLogMessage() string {
	return formatDuration(time.Now().Sub(logLastMessageRunTime))
}
func formatDuration(d time.Duration) string {
	d = d.Truncate(time.Millisecond)
	dur := durafmt.Parse(d).String()
	dur = strings.Replace(dur, " ", "", -1)

	{
		dur = strings.Replace(dur, "micros", "µs", 1)
		dur = strings.Replace(dur, "milliseconds", "ms", 1)
		dur = strings.Replace(dur, "seconds", "s", 1)
		dur = strings.Replace(dur, "minutes", "m", 1)
		dur = strings.Replace(dur, "hours", "h", 1)
		dur = strings.Replace(dur, "days", "D", 1)
		dur = strings.Replace(dur, "weeks", "W", 1)
		dur = strings.Replace(dur, "years", "Y", 1)
	}
	{
		dur = strings.Replace(dur, "millisecond", "ms", 1)
		dur = strings.Replace(dur, "second", "s", 1)
		dur = strings.Replace(dur, "minute", "m", 1)
		dur = strings.Replace(dur, "hour", "h", 1)
		dur = strings.Replace(dur, "day", "D", 1)
		dur = strings.Replace(dur, "week", "W", 1)
		dur = strings.Replace(dur, "year", "Y", 1)
	}
	return dur
}

// LogRunID adds to the log header the unique ID of this program run.
func LogRunID() string {
	return logRunID
}
func newParamsWithLogLevel(level string) []LogHeaderParameter {
	return []LogHeaderParameter{
		func() string {
			return level
		},
	}
}

var DefaultLogParameters = []LogHeaderParameter{
	//LogRunID,
	LogMessageNumber,
	LogParamTimestamp,
	//LogParamTimestampMs,
	LogElapsedFromStart,
	//LogElapsedFromLastLogMessage,
	//LogParamCallStack,
}

// Static:
var (
	logMu        = &sync.Mutex{}
	logStartedAt = time.Now()
	logRunID     = strings.ToUpper(RandomString(5))
)

// Dynamic:
var (
	logMessageCounter     int64
	logLastMessageRunTime time.Time
)

func DebugfWithParameters(params []LogHeaderParameter, format string, a ...interface{}) {
	header := getHeader(params)

	fmt.Fprintln(
		os.Stderr,
		header,
		fmt.Sprintf(
			format,
			a...,
		),
	)
}
func DebuglnWithParameters(params []LogHeaderParameter, a ...interface{}) {
	header := getHeader(params)

	fmt.Fprintln(
		os.Stderr,
		header,
		fmt.Sprintln(
			a...,
		),
	)
}
func getHeader(params []LogHeaderParameter) string {
	logMu.Lock()
	defer logMu.Unlock()

	var headerPrefix string
	{
		var headerVals []string
		for _, v := range params {
			headerVals = append(headerVals, v())
		}
		headerPrefix = strings.Join(headerVals, "|")
	}

	var headerVals []string
	for _, v := range DefaultLogParameters {
		headerVals = append(headerVals, v())
	}
	header := strings.Join(headerVals, "|")
	header = headerPrefix + "[" + header + "]"

	{
		logMessageCounter++
		logLastMessageRunTime = time.Now()
	}

	return header
}
func Debugf(format string, a ...interface{}) {
	DebugfWithParameters(
		newParamsWithLogLevel(DebugPrefix),
		format,
		a...,
	)
}
func Infof(format string, a ...interface{}) {
	DebugfWithParameters(
		newParamsWithLogLevel(InfoPrefix),
		format,
		a...,
	)
}
func Successf(format string, a ...interface{}) {
	DebugfWithParameters(
		newParamsWithLogLevel(SuccessPrefix),
		format,
		a...,
	)
}
func Warnf(format string, a ...interface{}) {
	DebugfWithParameters(
		newParamsWithLogLevel(WarnPrefix),
		format,
		a...,
	)
}

func Errorf(format string, a ...interface{}) {
	DebugfWithParameters(
		newParamsWithLogLevel(ErrorPrefix),
		RedBG(format),
		a...,
	)
}
func Fatalf(format string, a ...interface{}) {
	DebugfWithParameters(
		append(
			newParamsWithLogLevel(FatalPrefix),
			LogParamCallStack,
		),
		RedBG(format),
		a...,
	)
	os.Exit(1)
}
func Fataln(a ...interface{}) {
	DebuglnWithParameters(
		append(
			newParamsWithLogLevel(FatalPrefix),
			LogParamCallStack,
		),
		a...,
	)
	os.Exit(1)
}

func Black(s string) string {
	return rgbterm.FgString(s, 0, 0, 0)
}
func White(s string) string {
	return rgbterm.FgString(s, 255, 255, 255)
}
func BlackBG(s string) string {
	return rgbterm.BgString(s, 0, 0, 0)
}
func WhiteBG(s string) string {
	return Black(rgbterm.BgString(s, 255, 255, 255))
}
func Lime(str string) string {
	return rgbterm.FgString(str, 252, 255, 43)
}
func LimeBG(str string) string {
	return Black(rgbterm.BgString(str, 252, 255, 43))
}
func Yellow(message string) string {
	return tm.Color(message, tm.YELLOW)
}
func YellowBG(message string) string {
	return Black(tm.Background(message, tm.YELLOW))
}
func Orange(message string) string {
	return rgbterm.FgString(message, 255, 165, 0)
}
func OrangeBG(message string) string {
	return Black(rgbterm.BgString(message, 255, 165, 0))
}
func Red(str string) string {
	return rgbterm.FgString(str, 255, 0, 0)
}
func RedBG(s string) string {
	return tm.Color(tm.Background(s, tm.RED), tm.WHITE)
}

// light blue?
func Shakespeare(str string) string {
	return rgbterm.FgString(str, 82, 179, 217)
}
func ShakespeareBG(str string) string {
	return White(rgbterm.BgString(str, 82, 179, 217))
}

func Purple(s string) string {
	return rgbterm.FgString(s, 255, 0, 255)
}
func PurpleBG(s string) string {
	return Black(rgbterm.BgString(s, 255, 0, 255))
}
func Indigo(s string) string {
	return rgbterm.FgString(s, 75, 0, 130)
}
func IndigoBG(s string) string {
	return rgbterm.BgString(s, 75, 0, 130)
}

func Bold(message string) string {
	return tm.Bold(message)
}

func HighlightRedBG(str, substr string) string {
	return HighlightAnyCase(str, substr, RedBG)
}
func HighlightLimeBG(str, substr string) string {
	return HighlightAnyCase(str, substr, LimeBG)
}
func HighlightAnyCase(str, substr string, colorer func(string) string) string {
	substr = strings.ToLower(substr)
	str = strings.ToLower(str)

	hiSubstr := colorer(substr)
	return strings.Replace(str, substr, hiSubstr, -1)
}

func MoreThanOneIsTrue(bools ...bool) bool {
	var truthCount int
	for _, b := range bools {
		if b {
			truthCount++
		}
		if truthCount > 1 {
			return true
		}
	}
	return truthCount > 1
}
func AllTrue(bools ...bool) bool {
	for _, b := range bools {
		if !b {
			return false
		}
	}
	return true
}
func AllFalse(bools ...bool) bool {
	for _, b := range bools {
		if b {
			return false
		}
	}
	return true
}
func MustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(Sf("%s env var is not set", key))
	}
	return val
}

func StringToColor(str string) func(string) string {
	r, g, b, _ := calcColor(HashString(str))

	bgColor := WhiteBG
	if IsLight(r, g, b) {
		bgColor = BlackBG
	}
	return func(str string) string {
		return bgColor(rgbterm.FgString(str, uint8(r), uint8(g), uint8(b)))
	}
}
func StringToColorBG(str string) func(string) string {
	r, g, b, _ := calcColor(HashString(str))

	textColor := White
	if IsLight(r, g, b) {
		textColor = Black
	}
	return func(str string) string {
		return textColor(rgbterm.BgString(str, uint8(r), uint8(g), uint8(b)))
	}
}
func Colorize(str string) string {
	colorizer := StringToColor(str)
	return colorizer(str)
}
func ColorizeBG(str string) string {
	colorizer := StringToColorBG(str)
	return colorizer(str)
}
func calcColor(color uint64) (red, green, blue, alpha uint64) {
	alpha = color & 0xFF
	blue = (color >> 8) & 0xFF
	green = (color >> 16) & 0xFF
	red = (color >> 24) & 0xFF

	return red, green, blue, alpha
}

// IsLight returns whether the color is perceived to be a light color
func IsLight(rr, gg, bb uint64) bool {

	r := float64(rr)
	g := float64(gg)
	b := float64(bb)

	hsp := math.Sqrt(0.299*math.Pow(r, 2) + 0.587*math.Pow(g, 2) + 0.114*math.Pow(b, 2))

	return hsp > 130
}

func Dump(a ...interface{}) {
	spew.Dump(a...)
}
func Sdump(a ...interface{}) string {
	return spew.Sdump(a...)
}
func BellSound() {
	fmt.Print("\007")
}

type CombinedErrors struct {
	errs []error
}

//
func (ce *CombinedErrors) Error() string {
	buf := new(bytes.Buffer)
	buf.WriteString("The following errors occurred:")
	for _, err := range ce.errs {
		if err != nil {
			buf.WriteString("\n -  " + err.Error())
		}
	}
	return buf.String()
}
func allNil(errs ...error) bool {
	for _, err := range errs {
		if err != nil {
			return false
		}
	}
	return true
}
func CombineErrors(errs ...error) error {
	if len(errs) == 0 || allNil(errs...) {
		return nil
	}
	return &CombinedErrors{
		errs: errs,
	}
}

var hasherPool *sync.Pool

func init() {
	hasherPool = &sync.Pool{
		New: func() interface{} {
			return &maphash.Hash{}
		},
	}
}
func HashString(s string) uint64 {
	h := hasherPool.Get().(*maphash.Hash)

	defer hasherPool.Put(h)
	h.Reset()
	_, err := h.WriteString(s)
	if err != nil {
		panic(err)
	}

	return h.Sum64()
}
func HashBytes(b []byte) uint64 {
	h := hasherPool.Get().(*maphash.Hash)

	defer hasherPool.Put(h)
	h.Reset()
	_, err := h.Write(b)
	if err != nil {
		panic(err)
	}

	return h.Sum64()
}
func HashAnyWithJSON(v interface{}) (uint64, error) {
	// NOTE: this relies on the feature of json.Marshal that
	// sorts map keys in a constant way.
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	return HashBytes(b), nil
}

func MustHashAnyWithJSON(v interface{}) uint64 {
	h, err := HashAnyWithJSON(v)
	if err != nil {
		panic(err)
	}
	return h
}
