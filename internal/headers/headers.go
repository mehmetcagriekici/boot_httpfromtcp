package headers

import(
	"errors"
	"strings"
	"regexp"
	"fmt"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	crlf := strings.Index(str, "\r\n")
	if crlf == -1 {
		return 0, false, nil
	}

	if crlf == 0 {
		return len(str), true, nil
	}

	str = str[:crlf+2]
	harr := strings.Fields(strings.Trim(str, "\r\n"))
	if len(harr) == 1 {
		harr = append(harr, "")
	}
	
	if len(harr) != 2 {
		return 0, false, errors.New("Invalid Headers Struct")
	}

	fieldName := harr[0]
	if !strings.HasSuffix(fieldName, ":") {
		return 0, false, errors.New("Malformed Header")
	}
	fieldName = strings.TrimSuffix(fieldName, ":")
	pattern := "[a-zA-Z0-9!#$%&'*+-.^_`|~\\-]{1,}"
	re := regexp.MustCompile(pattern)
        matches := re.FindAllString(fieldName, -1)
	if len(matches) != 1 {
		return 0, false, errors.New("Invalid Fieldname")
	}
	fieldName = strings.ToLower(fieldName)
	
	if val, ok := h[fieldName]; ok {
		h[fieldName] = fmt.Sprintf("%s, %s", val, harr[1])
	} else {
		h[fieldName] = harr[1]
	}

	return len(str[:crlf+2]), false, nil
}
