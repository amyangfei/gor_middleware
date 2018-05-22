package gormw

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func invalidPayloadError(payload string) (string, error) {
	return "", fmt.Errorf("invalid payload: %s", payload)
}

func HttpMethod(payload string) (string, error) {
	pend := strings.IndexByte(payload, ' ')
	if pend == -1 {
		return invalidPayloadError(payload)
	}
	return payload[:pend], nil
}

func HttpPath(payload string) (string, error) {
	payload, err := url.PathUnescape(payload)
	if err != nil {
		return "", err
	}
	pstart := strings.IndexByte(payload, ' ')
	if pstart == -1 {
		return invalidPayloadError(payload)
	}
	pend := strings.IndexByte(payload[pstart+1:], ' ')
	if pend == -1 {
		return invalidPayloadError(payload)
	}
	return payload[pstart+1 : pstart+pend+1], nil
}

func SetHttpPath(payload, new_path string) (string, error) {
	pstart := strings.IndexByte(payload, ' ')
	if pstart == -1 {
		return invalidPayloadError(payload)
	}
	pend := strings.IndexByte(payload[pstart+1:], ' ')
	if pend == -1 {
		return invalidPayloadError(payload)
	}
	return payload[:pstart+1] + new_path + payload[pstart+pend+1:], nil
}

// Http response have status code in the same position as path for requests
func HttpStatus(payload string) (string, error) {
	return HttpPath(payload)
}

func SetHttpStatus(payload, new_status string) (string, error) {
	return SetHttpPath(payload, new_status)
}

func HttpPathParam(payload, name string) ([]string, error) {
	path, err := HttpPath(payload)
	if err != nil {
		return []string{}, err
	}
	u, err := url.Parse(path)
	if err != nil {
		return []string{}, err
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return []string{}, err
	}
	value, ok := m[name]
	if !ok {
		return []string{}, nil
	} else {
		return value, nil
	}
}

func SetHttpPathParam(payload, name, value string) (string, error) {
	pathQs, err := HttpPath(payload)
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile(name + "=([^&$]+)")
	newParam := name + "=" + value
	newPath := reg.ReplaceAllString(pathQs, newParam)
	if newPath == pathQs {
		if strings.IndexByte(newPath, '?') == -1 {
			newPath += "?"
		} else {
			newPath += "&"
		}
		newPath += name + "=" + value
	}
	return SetHttpPath(payload, url.PathEscape(newPath))
}

func HttpHeader(payload, name string) (map[string]interface{}, error) {
	currentLine := 0
	idx := 0
	header := map[string]interface{}{
		"start":  -1,
		"end":    -1,
		"vstart": -1,
	}
	for idx < len(payload) {
		c := payload[idx]
		if c == '\n' {
			currentLine += 1
			idx += 1
			header["end"] = idx
			start := header["start"].(int)
			vstart := header["vstart"].(int)
			if currentLine > 0 && start > 0 && vstart > 0 {
				key := payload[start : vstart-1]
				if strings.Compare(strings.ToLower(key), strings.ToLower(name)) == 0 {
					header["name"] = strings.ToLower(name)
					end := header["end"].(int)
					header["value"] = strings.TrimSpace(payload[vstart:end])
					return header, nil
				}
			}
			header["start"] = -1
			header["vstart"] = -1
		} else if c == '\r' {
			idx += 1
			continue
		} else if c == ':' {
			if header["vstart"].(int) == -1 {
				idx += 1
				header["vstart"] = idx
				continue
			}
		}
		if header["start"].(int) == -1 {
			header["start"] = idx
		}
		idx += 1
	}
	return nil, nil
}

func SetHttpHeader(payload, name, value string) (string, error) {
	header, err := HttpHeader(payload, name)
	if err != nil {
		return "", err
	}
	if header == nil {
		header_start := strings.IndexByte(payload, '\n') + 1
		if header_start == 0 {
			return invalidPayloadError(payload)
		}
		return payload[:header_start] + name + ": " + value + "\r\n" + payload[header_start:], nil
	} else {
		return payload[:header["vstart"].(int)] + " " + value + "\r\n" + payload[header["end"].(int):], nil
	}
}
