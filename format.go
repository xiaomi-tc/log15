package log15

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/xiaomi-tc/log15/structured"
)

const (
	//timeFormat     = "2006-01-02T15:04:05-0700"
	timeFormat     = "2006/01/02 15:04:05"
	termTimeFormat = "01-02|15:04:05"
	floatFormat    = 'f'
	termMsgJust    = 40

	// DurationFieldUnit defines the unit for time.Duration type fields added
	// using the Dur method.
	DurationFieldUnit = time.Millisecond

	// DurationFieldInteger renders Dur fields as integer instead of float if
	// set to true.
	DurationFieldInteger = false
)

var (
	enc     structured.Encoder
	bufPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 256)
		},
	}
)

type Format interface {
	Format(r *Record) []byte
}

// FormatFunc returns a new Format object which uses
// the given function to perform record formatting.
func FormatFunc(f func(*Record) []byte) Format {
	return formatFunc(f)
}

type formatFunc func(*Record) []byte

func (f formatFunc) Format(r *Record) []byte {
	return f(r)
}

// TerminalFormat formats log records optimized for human readability on
// a terminal with color-coded level output and terser human friendly timestamp.
// This format should only be used for interactive programs or while developing.
//
//     [TIME] [LEVEL] MESAGE key=value key=value ...
//
// Example:
//
//     [May 16 20:58:45] [DBUG] remove route ns=haproxy addr=127.0.0.1:50002
//
func TerminalFormat() Format {
	return FormatFunc(func(r *Record) []byte {
		var color = 0
		switch r.Lvl {
		case LvlCrit:
			color = 35
		case LvlError:
			color = 31
		case LvlWarn:
			color = 33
		case LvlInfo:
			color = 32
		case LvlDebug:
			color = 36
		}

		var buf = make([]byte, 0, 256)
		lvl := strings.ToUpper(r.Lvl.String())

		// head
		buf = appendColordString(buf, lvl, color)
		buf = append(buf, ' ')

		buf = append(buf, '[')
		buf = enc.AppendTime(buf, r.Time, timeFormat)
		buf = append(buf, "] ["...)
		buf = append(buf, []byte(r.Call)...)
		buf = append(buf, ']')
		if r.RequestID != "" {
			buf = append(buf, " ["...)
			buf = append(buf, []byte(r.KeyNames.ReqID)...)
			buf = append(buf, '=')
			buf = append(buf, []byte(r.RequestID)...)
			buf = append(buf, ']')
		}

		// msg
		buf = append(buf, ' ')
		buf = appendColordString(buf, r.KeyNames.Msg, color)
		buf = append(buf, '=')
		buf = enc.AppendString(buf, r.Msg)

		// fields
		buf = logfmt(buf, r.Ctx, color)

		return buf
	})
}

// LogfmtFormat prints records in logfmt format, an easy machine-parseable but human-readable
// format for key/value pairs.
//
// For more details see: http://godoc.org/github.com/kr/logfmt
//
func LogfmtFormat() Format {
	return FormatFunc(func(r *Record) []byte {
		// assignment in serial processing
		logLevel = byte(r.Lvl)
		logMetaKey = r.MetaK
		logMetaValue = r.MetaV

		var caller string
		if r.CustomCaller == "" {
			caller = r.Call
		} else {
			caller = r.CustomCaller
		}

		var buf = make([]byte, 0, 256)

		// log head
		buf = append(buf, '[')
		buf = enc.AppendTime(buf, r.Time, timeFormat)
		buf = append(buf, "] ["...)
		buf = append(buf, []byte(r.Lvl.String())...)
		buf = append(buf, "] ["...)
		buf = append(buf, []byte(caller)...)
		buf = append(buf, "] "...)
		if r.RequestID != "" {
			buf = append(buf, " ["...)
			buf = append(buf, []byte(r.KeyNames.ReqID)...)
			buf = append(buf, '=')
			buf = append(buf, []byte(r.RequestID)...)
			buf = append(buf, "] "...)
		}

		// msg
		buf = appendFields(buf, r.KeyNames.Msg, r.Msg)

		// fields
		buf = logfmt(buf, r.Ctx, 0)

		return buf
	})
}

func logfmt(buf []byte, ctx []interface{}, color int) []byte {
	var sz = len(ctx)
	for i := 0; i < sz; i += 2 {
		buf = append(buf, ' ')

		k, ok := ctx[i].(string)
		v := ctx[i+1]
		if !ok {
			k, v = errorKey, k
		}

		buf = appendColordString(buf, k, color)
		buf = append(buf, '=')
		buf = appendVal(buf, v)
	}

	buf = append(buf, '\n')
	return buf
}

// JsonFormat formats log records as JSON objects separated by newlines.
// It is the equivalent of JsonFormatEx(false, true).
func JsonFormat() Format {
	return JsonFormatEx(false, true)
}

// JsonFormatEx formats log records as JSON objects. If pretty is true,
// records will be pretty-printed. If lineSeparated is true, records
// will be logged with a new line between each record.
func JsonFormatEx(pretty, lineSeparated bool) Format {
	jsonMarshal := json.Marshal
	if pretty {
		jsonMarshal = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		}
	}

	return FormatFunc(func(r *Record) []byte {
		props := make(map[string]interface{})

		props[r.KeyNames.Time] = r.Time
		props[r.KeyNames.Lvl] = r.Lvl.String()
		props[r.KeyNames.Msg] = r.Msg

		for i := 0; i < len(r.Ctx); i += 2 {
			k, ok := r.Ctx[i].(string)
			if !ok {
				props[errorKey] = fmt.Sprintf("%+v is not a string key", r.Ctx[i])
			}
			props[k] = formatJsonValue(r.Ctx[i+1])
		}

		b, err := jsonMarshal(props)
		if err != nil {
			b, _ = jsonMarshal(map[string]string{
				errorKey: err.Error(),
			})
			return b
		}

		if lineSeparated {
			b = append(b, '\n')
		}

		return b
	})
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}

func formatJsonValue(value interface{}) interface{} {
	value = formatShared(value)
	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}

// formatValue formats a value for serialization
func formatLogfmtValue(value interface{}) string {
	var buf = make([]byte, 0, 32)
	buf = appendVal(buf, value)
	return string(buf)
}

func appendColordString(dst []byte, k string, color int) []byte {
	if color > 0 {
		dst = append(dst, "\x1b["...)
		dst = enc.AppendInt(dst, color)
		dst = append(dst, 'm')
		dst = append(dst, []byte(k)...)
		dst = append(dst, "\x1b[0m"...)
	} else {
		dst = append(dst, []byte(k)...)
	}
	return dst
}

func appendFields(dst []byte, key string, val interface{}) []byte {
	dst = enc.AppendKey(dst, key)
	dst = appendVal(dst, val)
	return dst
}

func appendVal(dst []byte, val interface{}) []byte {
	switch val := val.(type) {
	case string:
		dst = enc.AppendString(dst, val)
	case []byte:
		dst = enc.AppendBytes(dst, val)
	case error:
		dst = enc.AppendString(dst, val.Error())
	case []error:
		dst = enc.AppendArrayStart(dst)
		for i, err := range val {
			dst = enc.AppendString(dst, err.Error())

			if i < (len(val) - 1) {
				enc.AppendArrayDelim(dst)
			}
		}
		dst = enc.AppendArrayEnd(dst)
	case bool:
		dst = enc.AppendBool(dst, val)
	case int:
		dst = enc.AppendInt(dst, val)
	case int8:
		dst = enc.AppendInt8(dst, val)
	case int16:
		dst = enc.AppendInt16(dst, val)
	case int32:
		dst = enc.AppendInt32(dst, val)
	case int64:
		dst = enc.AppendInt64(dst, val)
	case uint:
		dst = enc.AppendUint(dst, val)
	case uint8:
		dst = enc.AppendUint8(dst, val)
	case uint16:
		dst = enc.AppendUint16(dst, val)
	case uint32:
		dst = enc.AppendUint32(dst, val)
	case uint64:
		dst = enc.AppendUint64(dst, val)
	case float32:
		dst = enc.AppendFloat32(dst, val)
	case float64:
		dst = enc.AppendFloat64(dst, val)
	case time.Time:
		dst = enc.AppendTime(dst, val, timeFormat)
	case time.Duration:
		dst = enc.AppendDuration(dst, val, DurationFieldUnit, DurationFieldInteger)
	case *string:
		if val != nil {
			dst = enc.AppendString(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *bool:
		if val != nil {
			dst = enc.AppendBool(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *int:
		if val != nil {
			dst = enc.AppendInt(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *int8:
		if val != nil {
			dst = enc.AppendInt8(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *int16:
		if val != nil {
			dst = enc.AppendInt16(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *int32:
		if val != nil {
			dst = enc.AppendInt32(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *int64:
		if val != nil {
			dst = enc.AppendInt64(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *uint:
		if val != nil {
			dst = enc.AppendUint(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *uint8:
		if val != nil {
			dst = enc.AppendUint8(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *uint16:
		if val != nil {
			dst = enc.AppendUint16(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *uint32:
		if val != nil {
			dst = enc.AppendUint32(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *uint64:
		if val != nil {
			dst = enc.AppendUint64(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *float32:
		if val != nil {
			dst = enc.AppendFloat32(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *float64:
		if val != nil {
			dst = enc.AppendFloat64(dst, *val)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *time.Time:
		if val != nil {
			dst = enc.AppendTime(dst, *val, timeFormat)
		} else {
			dst = enc.AppendNil(dst)
		}
	case *time.Duration:
		if val != nil {
			dst = enc.AppendDuration(dst, *val, DurationFieldUnit, DurationFieldInteger)
		} else {
			dst = enc.AppendNil(dst)
		}
	case []string:
		dst = enc.AppendStrings(dst, val)
	case []bool:
		dst = enc.AppendBools(dst, val)
	case []int:
		dst = enc.AppendInts(dst, val)
	case []int8:
		dst = enc.AppendInts8(dst, val)
	case []int16:
		dst = enc.AppendInts16(dst, val)
	case []int32:
		dst = enc.AppendInts32(dst, val)
	case []int64:
		dst = enc.AppendInts64(dst, val)
	case []uint:
		dst = enc.AppendUints(dst, val)
	// case []uint8:
	// 	dst = enc.AppendUints8(dst, val)
	case []uint16:
		dst = enc.AppendUints16(dst, val)
	case []uint32:
		dst = enc.AppendUints32(dst, val)
	case []uint64:
		dst = enc.AppendUints64(dst, val)
	case []float32:
		dst = enc.AppendFloats32(dst, val)
	case []float64:
		dst = enc.AppendFloats64(dst, val)
	case []time.Time:
		dst = enc.AppendTimes(dst, val, timeFormat)
	case []time.Duration:
		dst = enc.AppendDurations(dst, val, DurationFieldUnit, DurationFieldInteger)
	case nil:
		dst = enc.AppendNil(dst)
	case net.IP:
		dst = enc.AppendIPAddr(dst, val)
	case net.IPNet:
		dst = enc.AppendIPPrefix(dst, val)
	case net.HardwareAddr:
		dst = enc.AppendMACAddr(dst, val)
	default:
		dst = enc.AppendInterface(dst, val)
	}

	return dst
}
