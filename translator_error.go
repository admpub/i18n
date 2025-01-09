package i18n

import "errors"

// translatorError implements the error interface for use in this package. it
// keeps an optional reference to a Translator instance, which it uses to
// include which locale the error occurs with in the error message returned by
// the Error() method
type translatorError struct {
	translator *Translator
	message    string
	rawError   error
}

// Error satisfies the error interface requirements
func (e translatorError) Error() string {
	if e.translator != nil {
		return "translator error (locale: " + e.translator.locale + ") - " + e.message
	}
	return "translator error - " + e.message
}

func (e translatorError) Unwrap() error {
	return e.rawError
}

var (
	ErrKeyNotFound               = errors.New(`key not found`)
	ErrUnknownDatetimeFormat     = errors.New(`unknown datetime format`)
	ErrUnknownDatetimeFormatUnit = errors.New(`unknown datetime format unit`)
	ErrUnsupportedYearLength     = errors.New(`unsupported year length`)
	ErrUnsupportedMonthLength    = errors.New(`unsupported month length`)
	ErrUnsupportedYearDayofweek  = errors.New(`unsupported year day-of-week`)
	ErrUnsupportedDayofweek      = errors.New(`unsupported day-of-year`)
	ErrUnsupportedHour12         = errors.New(`unsupported hour-12`)
	ErrUnsupportedHour24         = errors.New(`unsupported hour-24`)
	ErrUnsupportedMinute         = errors.New(`unsupported minute`)
	ErrUnsupportedSecond         = errors.New(`unsupported second`)
	ErrUnsupportedDayPeriod      = errors.New(`unsupported day-period`)
	ErrMalformedDatetimeFormat   = errors.New(`malformed datetime format`)
)
