package cast

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const UTC = "UTC"

func ToDuration(a any) time.Duration {
	v, _ := ToDurationE(a)
	return v
}

func ToDurationE(a any) (time.Duration, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case string:
		return stringToDurationE(v)
	default:
		return 0, fmt.Errorf("invalid duration type: %T", a)
	}
}

func ToTimeZone(a any) *time.Location {
	v, _ := ToTimeZoneE(a)
	return v
}

func ToTimeZoneE(a any) (*time.Location, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case string:
		return stringToTimeZoneE(v)
	default:
		return nil, fmt.Errorf("invalid time zone type: %T", a)
	}
}

var (
	durationRegExp       *regexp.Regexp
	durationRegExpGroups = []string{
		`<years>[\+|\-]?\d+[Y|y]`,
		`<months>[\+|\-]?\d+M`,
		`<days>[\+|\-]?\d+[D|d]`,
		`<hours>[\+|\-]?\d+[H|h]`,
		`<minutes>[\+|\-]?\d+m`,
		`<seconds>[\+|\-]?\d+[S|s]`,
	}
)

func init() {
	var buf = new(bytes.Buffer)
	for _, group := range durationRegExpGroups {
		buf.WriteString(`(?P`)
		buf.WriteString(group)
		buf.WriteString(`)?`)
	}
	durationRegExp = regexp.MustCompile(buf.String())
}

func stringToDurationE(str string) (time.Duration, error) {
	lastChar := str[len(str)-1]
	if lastChar >= '0' && lastChar <= '9' {
		str += "s"
	}

	matches := durationRegExp.FindStringSubmatch(str)

	if len(matches) == 0 {
		return 0, fmt.Errorf("parse duration `%s` failed, empty match", str)
	}

	nums := []int{}
	for index := 1; index < len(matches); index++ {
		s := matches[index]
		if len(s) == 0 {
			nums = append(nums, 0)
			continue
		}
		for s[len(s)-1] < '0' || s[len(s)-1] > '9' {
			s = s[:len(s)-1]
		}
		n, err := ToInt64E(s)
		if err != nil {
			return 0, fmt.Errorf("parse duration `%s` failed, %v", str, err)
		}
		nums = append(nums, int(n))
	}

	// TODO: to support time format str as tm value
	tm := time.Now().UTC()
	duration := tm.AddDate(nums[0], nums[1], nums[2]).Add(
		time.Duration(nums[3]) * time.Hour,
	).Add(
		time.Duration(nums[4]) * time.Minute,
	).Add(
		time.Duration(nums[5]) * time.Second,
	).Sub(tm)

	return duration, nil
}

func stringToTimeZoneE(s string) (*time.Location, error) {
	// Check first char is + or -, or is digit
	if s[0] == '+' || s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
		duration, err := stringToDurationE(s)
		if err != nil {
			return nil, err
		}
		return durationToLocation(duration), nil
	} else if strings.HasPrefix(s, UTC) {
		duration, err := stringUTCToDurationE(s)
		if err != nil {
			return nil, err
		}
		return durationToLocation(duration), nil
	} else {
		// Check timezone is valid
		if loc, err := time.LoadLocation(s); err != nil {
			return nil, err
		} else {
			// get time zone offset
			_, offset := time.Now().In(loc).Zone()
			duration := ToDuration(offset)
			return durationToLocation(duration), nil
		}
	}
}

func durationToStringUTC(duration time.Duration) string {
	seconds := int(duration.Seconds())
	if seconds < 0 {
		h := -seconds / 3600
		m := -seconds % 3600 / 60
		s := -seconds % 3600 % 60
		str := UTC
		if h != 0 {
			str += fmt.Sprintf("-%02d", h)
		}
		if m != 0 {
			str += fmt.Sprintf(":%02d", m)
		}
		if s != 0 {
			str += fmt.Sprintf(":%02d", s)
		}
		return str
	} else {
		h := seconds / 3600
		m := seconds % 3600 / 60
		s := seconds % 3600 % 60
		str := UTC
		if h != 0 {
			str += fmt.Sprintf("+%02d", h)
		}
		if m != 0 {
			str += fmt.Sprintf(":%02d", m)
		}
		if s != 0 {
			str += fmt.Sprintf(":%02d", s)
		}
		return str
	}
}

func durationToLocation(duration time.Duration) *time.Location {
	return time.FixedZone(durationToStringUTC(duration), int(duration.Seconds()))
}

func stringUTCToDurationE(name string) (time.Duration, error) {
	var sign = 1

	if name == UTC {
		return 0, nil
	} else if strings.HasPrefix(name, "UTC+") {
		name = name[4:]
	} else if strings.HasPrefix(name, "UTC-") {
		name = name[4:]
		sign = -1
	} else {
		return 0, fmt.Errorf("invalid timezone name `%s`", name)
	}

	parts := strings.Split(name, ":")
	if len(parts) == 1 {
		h, err := ToInt64E(parts[0])
		if err != nil {
			return 0, err
		}
		return time.Duration(sign) * time.Duration(h) * time.Hour, nil
	} else if len(parts) == 2 {
		h, err := ToInt64E(parts[0])
		if err != nil {
			return 0, err
		}
		m, err := ToInt64E(parts[1])
		if err != nil {
			return 0, err
		}
		return time.Duration(sign) * (time.Duration(h)*time.Hour + time.Duration(m)*time.Minute), nil
	} else if len(parts) == 3 {
		h, err := ToInt64E(parts[0])
		if err != nil {
			return 0, err
		}
		m, err := ToInt64E(parts[1])
		if err != nil {
			return 0, err
		}
		s, err := ToInt64E(parts[2])
		if err != nil {
			return 0, err
		}
		return time.Duration(sign) * (time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second), nil
	} else {
		return 0, fmt.Errorf("invalid time zone name `%s`", name)
	}
}

// func ParseTimeZone(s string) int64 {
// 	// Check first char is + or -, or is digit
// 	if s[0] == '+' || s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
// 		duration := ParseDuration(s, time.Now())
// 		return ToInt64(duration.Seconds())
// 	} else {
// 		// Check timezone is valid
// 		if loc, err := time.LoadLocation(s); err != nil {
// 			panic(err)
// 		} else {
// 			// get time zone offset
// 			_, offset := time.Now().In(loc).Zone()
// 			return ToInt64(offset)
// 		}
// 	}
// }

// TODO support time & duration

// // ToTime casts an interface to a time.Time type.
// func ToTime(a any, args ...any) time.Time {
// 	t, _ := ToTimeE(a)
// 	return t
// }

// // ToTimeE casts an interface to a time.Time type.
// func ToTimeE(a any, args ...any) (time.Time, error) {
// 	a = indirectToStringerOrError(a)

// 	switch v := a.(type) {
// 	case int:
// 		return time.Unix(int64(v), 0), nil
// 	case int8:
// 		return time.Unix(int64(v), 0), nil
// 	case int16:
// 		return time.Unix(int64(v), 0), nil
// 	case int32:
// 		return time.Unix(int64(v), 0), nil
// 	case int64:
// 		return time.Unix(v, 0), nil
// 	case uint:
// 		return time.Unix(int64(v), 0), nil
// 	case uint8:
// 		return time.Unix(int64(v), 0), nil
// 	case uint16:
// 		return time.Unix(int64(v), 0), nil
// 	case uint32:
// 		return time.Unix(int64(v), 0), nil
// 	case uint64:
// 		return time.Unix(int64(v), 0), nil
// 	case float32:
// 		return time.Unix(int64(v), 0), nil
// 	case float64:
// 		return time.Unix(int64(v), 0), nil
// 	case *big.Int:
// 		return time.Unix(v.Int64(), 0), nil
// 	case *big.Float:
// 		n, _ := v.Int64()
// 		return time.Unix(n, 0), nil
// 	case *big.Rat:
// 		n, _ := v.Float64()
// 		return time.Unix(int64(n), 0), nil
// 	case complex64:
// 		return time.Unix(int64(real(v)), 0), nil
// 	case complex128:
// 		return time.Unix(int64(real(v)), 0), nil
// 	case bool:
// 		if v {
// 			return time.Unix(1, 0), nil
// 		}
// 		return time.Unix(0, 0), nil
// 	case string:
// 		return parseTime(v, args...)
// 	case []byte:
// 		return parseTime(string(v), args...)
// 	case fmt.Stringer:
// 		return parseTime(v.String(), args...)
// 	case nil:
// 		return time.Time{}, nil
// 	default:
// 		return time.Time{}, fmt.Errorf("unable to cast %#v of type %T to Time", a, a)
// 	}
// }

// // ToDuration casts an interface to a time.Duration type.
// func ToDuration(i any, args ...any) time.Duration {
// 	d, _ := ToDurationE(i)
// 	return d
// }

// // ToDurationE casts an interface to a time.Duration type.
// func ToDurationE(a any, args ...any) (d time.Duration, err error) {
// 	a = indirectToStringerOrError(a)

// 	switch v := a.(type) {
// 	case int:
// 		return time.Duration(v), nil
// 	case int8:
// 		return time.Duration(v), nil
// 	case int16:
// 		return time.Duration(v), nil
// 	case int32:
// 		return time.Duration(v), nil
// 	case int64:
// 		return time.Duration(v), nil
// 	case uint:
// 		return time.Duration(v), nil
// 	case uint8:
// 		return time.Duration(v), nil
// 	case uint16:
// 		return time.Duration(v), nil
// 	case uint32:
// 		return time.Duration(v), nil
// 	case uint64:
// 		return time.Duration(v), nil
// 	case float32:
// 		return time.Duration(v), nil
// 	case float64:
// 		return time.Duration(v), nil
// 	case *big.Int:
// 		return time.Duration(v.Int64()), nil
// 	case *big.Float:
// 		n, _ := v.Int64()
// 		return time.Duration(n), nil
// 	case *big.Rat:
// 		n, _ := v.Float64()
// 		return time.Duration(n), nil
// 	case complex64:
// 		return time.Duration(real(v)), nil
// 	case complex128:
// 		return time.Duration(real(v)), nil
// 	case bool:
// 		if v {
// 			return time.Duration(1), nil
// 		}
// 		return time.Duration(0), nil
// 	case string:
// 		return parseDuration(v, args...)
// 	case []byte:
// 		return parseDuration(string(v), args...)
// 	case fmt.Stringer:
// 		return parseDuration(v.String(), args...)
// 	case error:
// 		return parseDuration(v.Error(), args...)
// 	default:
// 		return time.Duration(0), fmt.Errorf("unable to cast %#v of type %T to Duration", a, a)
// 	}
// }

// type TimeFormatType int

// const (
// 	TimeFormatNoTimezone TimeFormatType = iota
// 	TimeFormatNamedTimezone
// 	TimeFormatNumericTimezone
// 	TimeFormatNumericAndNamedTimezone
// 	TimeFormatTimeOnly
// )

// func _() {
// 	// An "invalid array index" compiler error signifies that the constant values have changed.
// 	// Re-run the stringer command to generate them again.
// 	var x [1]struct{}
// 	_ = x[TimeFormatNoTimezone-0]
// 	_ = x[TimeFormatNamedTimezone-1]
// 	_ = x[TimeFormatNumericTimezone-2]
// 	_ = x[TimeFormatNumericAndNamedTimezone-3]
// 	_ = x[TimeFormatTimeOnly-4]
// }

// const _timeFormatType_name = "timeFormatNoTimezonetimeFormatNamedTimezonetimeFormatNumericTimezonetimeFormatNumericAndNamedTimezonetimeFormatTimeOnly"

// var _timeFormatType_index = [...]uint8{0, 20, 43, 68, 101, 119}

// func (i TimeFormatType) String() string {
// 	if i < 0 || i >= TimeFormatType(len(_timeFormatType_index)-1) {
// 		return "timeFormatType(" + strconv.FormatInt(int64(i), 10) + ")"
// 	}
// 	return _timeFormatType_name[_timeFormatType_index[i]:_timeFormatType_index[i+1]]
// }

// type TimeFormat struct {
// 	Format string
// 	Type   TimeFormatType
// }

// func (f TimeFormat) hasTimezone() bool {
// 	// We don't include the formats with only named timezones, see
// 	// https://github.com/golang/go/issues/19694#issuecomment-289103522
// 	return f.Type >= TimeFormatNumericTimezone && f.Type <= TimeFormatNumericAndNamedTimezone
// }

// var (
// 	defaultTimeFormats = []TimeFormat{
// 		{time.RFC3339, TimeFormatNumericTimezone},
// 		{"2006-01-02T15:04:05", TimeFormatNoTimezone}, // iso8601 without timezone
// 		{time.RFC1123Z, TimeFormatNumericTimezone},
// 		{time.RFC1123, TimeFormatNamedTimezone},
// 		{time.RFC822Z, TimeFormatNumericTimezone},
// 		{time.RFC822, TimeFormatNamedTimezone},
// 		{time.RFC850, TimeFormatNamedTimezone},
// 		{"2006-01-02 15:04:05.999999999 -0700 MST", TimeFormatNumericAndNamedTimezone}, // Time.String()
// 		{"2006-01-02T15:04:05-0700", TimeFormatNumericTimezone},                        // RFC3339 without timezone hh:mm colon
// 		{"2006-01-02 15:04:05Z0700", TimeFormatNumericTimezone},                        // RFC3339 without T or timezone hh:mm colon
// 		{"2006-01-02 15:04:05", TimeFormatNoTimezone},
// 		{time.ANSIC, TimeFormatNoTimezone},
// 		{time.UnixDate, TimeFormatNamedTimezone},
// 		{time.RubyDate, TimeFormatNumericTimezone},
// 		{"2006-01-02 15:04:05Z07:00", TimeFormatNumericTimezone},
// 		{"2006-01-02", TimeFormatNoTimezone},
// 		{"02 Jan 2006", TimeFormatNoTimezone},
// 		{"2006-01-02 15:04:05 -07:00", TimeFormatNumericTimezone},
// 		{"2006-01-02 15:04:05 -0700", TimeFormatNumericTimezone},
// 		{time.Kitchen, TimeFormatTimeOnly},
// 		{time.Stamp, TimeFormatTimeOnly},
// 		{time.StampMilli, TimeFormatTimeOnly},
// 		{time.StampMicro, TimeFormatTimeOnly},
// 		{time.StampNano, TimeFormatTimeOnly},
// 	}

// 	location = time.UTC
// )

// func parseTime(s string, args ...any) (d time.Time, e error) {
// 	var (
// 		location    *time.Location
// 		timeFormats []TimeFormat
// 	)
// 	for _, arg := range args {
// 		switch v := arg.(type) {
// 		case *time.Location:
// 			location = v
// 		case TimeFormat:
// 			timeFormats = append(timeFormats, v)
// 		}
// 	}

// 	if location == nil {
// 		location = time.Local
// 	}
// 	if len(timeFormats) == 0 {
// 		timeFormats = defaultTimeFormats
// 	}

// 	for _, timeFormat := range timeFormats {
// 		if d, e = time.Parse(timeFormat.Format, s); e == nil {

// 			// Some time formats have a zone name, but no offset, so it gets
// 			// put in that zone name (not the default one passed in to us), but
// 			// without that zone's offset. So set the location manually.
// 			if timeFormat.Type <= TimeFormatNamedTimezone {
// 				if location == nil {
// 					location = time.Local
// 				}
// 				year, month, day := d.Date()
// 				hour, min, sec := d.Clock()
// 				d = time.Date(year, month, day, hour, min, sec, d.Nanosecond(), location)
// 			}

// 			return
// 		}
// 	}
// 	return d, fmt.Errorf("unable to parse date: %s", s)
// }

// var (
// 	durationRegExp       *regexp.Regexp
// 	durationRegExpGroups = []string{
// 		`<years>[\+|\-]?\d+y`,
// 		`<months>[\+|\-]?\d+M`,
// 		`<days>[\+|\-]?\d+d`,
// 		`<hours>[\+|\-]?\d+h`,
// 		`<minutes>[\+|\-]?\d+m`,
// 		`<seconds>[\+|\-]?\d+s`,
// 		`<milliseconds>[\+|\-]?\d+ms`,
// 		`<microseconds>[\+|\-]?\d+us`,
// 		`<nanoseconds>[\+|\-]?\d+ns`,
// 	}
// )

// func init() {
// 	var buf = new(bytes.Buffer)
// 	for _, group := range durationRegExpGroups {
// 		buf.WriteString(`(?P`)
// 		buf.WriteString(group)
// 		buf.WriteString(`)?`)
// 	}
// 	durationRegExp = regexp.MustCompile(buf.String())
// }

// func parseDuration(s string, args ...any) (time.Duration, error) {
// 	var (
// 		t = time.Now()
// 	)

// 	for _, arg := range args {
// 		switch v := arg.(type) {
// 		case time.Time:
// 			t = v
// 		}
// 	}

// 	matches := durationRegExp.FindStringSubmatch(s)
// 	if len(matches) == 0 {
// 		return 0, nil
// 	}

// 	nums := []int{}
// 	for index := 1; index < len(matches); index++ {
// 		s := matches[index]
// 		if len(s) == 0 {
// 			nums = append(nums, 0)
// 			continue
// 		}
// 		for s[len(s)-1] < '0' || s[len(s)-1] > '9' {
// 			s = s[:len(s)-1]
// 		}
// 		n, err := strconv.ParseInt(s, 10, 64)
// 		if err != nil {
// 			return 0, err
// 		}
// 		nums = append(nums, int(n))
// 	}

// 	if len(nums) == 0 {
// 		return 0, fmt.Errorf("parse duration `%s` failed", s)
// 	}

// 	duration := t.AddDate(nums[0], nums[1], nums[2]).Add(
// 		time.Duration(nums[3]) * time.Hour,
// 	).Add(
// 		time.Duration(nums[4]) * time.Minute,
// 	).Add(
// 		time.Duration(nums[5]) * time.Second,
// 	).Add(
// 		time.Duration(nums[6]) * time.Millisecond,
// 	).Add(
// 		time.Duration(nums[7]) * time.Microsecond,
// 	).Add(
// 		time.Duration(nums[8]) * time.Nanosecond,
// 	).Sub(t)

// 	return duration, nil
// }
