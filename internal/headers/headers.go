package headers

import(
	"errors"
	"strings"
	"regexp"
	"fmt"
)

type Headers map[string]string

func (h Headers) Get(key string) (string, bool) {
	if val, ok := h[key]; ok {
		return val, true
	}
	return "", false
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	
	cr := strings.Index(str, "\r\n")
	if cr == -1 {
		return 0, false, nil
	}

	if cr == 0 {
		return 2, true, nil
	}

	str = str[:cr+2]
	harr := strings.Fields(strings.Trim(str, "\r\n"))
	if len(harr) == 0 {
		return 0, true, nil
	}
	
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

	return len(str[:cr+2]), false, nil
}
