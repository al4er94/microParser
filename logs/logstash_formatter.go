package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"sort"
	"sync"
)

type LogstashMessageFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force disabling colors. For a TTY colors are enabled by default.
	DisableColors bool

	// Force formatted layout, even for non-TTY output.
	ForceFormatting bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Disable the conversion of the log levels to uppercase
	DisableUppercase bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

	// Timestamp format to use for display when a full timestamp is printed.
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// Wrap empty fields in quotes if true.
	QuoteEmptyFields bool

	// Can be set to the override the default quoting character "
	// with something else. For example: ', or `.
	QuoteCharacter string

	// Pad msg field with spaces on the right for display.
	// The value for this parameter will be the size of padding.
	// Its default value is zero, which means no padding will be applied for msg.
	SpacePadding int

	sync.Once
}

func (f *LogstashMessageFormatter) init(entry *logrus.Entry) {
	if len(f.QuoteCharacter) == 0 {
		f.QuoteCharacter = "\""
	}
}

func (f *LogstashMessageFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buffer *bytes.Buffer

	keys := make([]string, len(entry.Data))
	keysIndex := 0

	for k := range entry.Data {
		keys[keysIndex] = k
		keysIndex++
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}

	if entry.Buffer != nil {
		buffer = entry.Buffer
	} else {
		buffer = &bytes.Buffer{}
	}

	prefixFieldClashes(entry.Data)

	f.Do(func() { f.init(entry) })

	f.printColored(buffer, entry)

	buffer.WriteByte('\n')

	return buffer.Bytes(), nil
}

func (f *LogstashMessageFormatter) printColored(
	buffer *bytes.Buffer,
	entry *logrus.Entry,
) {
	message := entry.Message

	messageFormat := "%s"

	if f.SpacePadding != 0 {
		messageFormat = fmt.Sprintf("%%-%ds", f.SpacePadding)
	}

	fmt.Fprintf(buffer, messageFormat, message)
}

// This is to not silently overwrite `time`, `msg` and `level` fields when
// dumping it. If this code wasn't there doing:
//
//	logrus.WithField("level", 1).Info("hello")
//
// would just silently drop the user provided level. Instead with this code
// it'll be logged as:
//
//	{"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
func prefixFieldClashes(data logrus.Fields) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
	}

	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}

	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}
}
