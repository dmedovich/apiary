package annotation

import (
	"mime"
	"strconv"
	"strings"
)

type ErrorSpec struct {
	Code   int
	Schema string
}

type Operation struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Errors      []ErrorSpec

	Security []string

	Request             string
	Response            string
	ContentType         string
	ResponseStatus      int
	ResponseContentType string
	ResponseFormat      string

	Warnings []string
}

func looksLikeKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && r != '-' {
			return false
		}
	}
	return true
}

func Parse(lines []string) (*Operation, bool) {
	op := &Operation{}
	found := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if rest, ok := strings.CutPrefix(line, "apiary:operation "); ok {
			parts := strings.Fields(rest)
			if len(parts) < 2 {
				continue
			}
			op.Method = strings.ToUpper(parts[0])
			op.Path = parts[1]
			found = true
			continue
		}

		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		switch key {
		case "operationId":
			op.OperationID = value
		case "summary":
			op.Summary = value
		case "description":
			op.Description = value
		case "tags":
			for _, tag := range strings.Split(value, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					op.Tags = append(op.Tags, tag)
				}
			}
		case "errors":
			for _, item := range strings.Split(value, ",") {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				parts := strings.Fields(item)
				n, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				spec := ErrorSpec{Code: n}
				if len(parts) > 1 {
					spec.Schema = parts[1]
				}
				op.Errors = append(op.Errors, spec)
			}
		case "security":
			for _, s := range strings.Split(value, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					op.Security = append(op.Security, s)
				}
			}
		case "request":
			op.Request = value
		case "response":
			op.Response = value
		case "response-status":
			code, err := strconv.Atoi(value)
			if err != nil || code < 200 || code > 299 {
				op.Warnings = append(op.Warnings, "invalid response-status \""+value+"\"; expected a success status from 200 to 299; ignored")
				continue
			}
			op.ResponseStatus = code
		case "response-content-type":
			mediaType, _, err := mime.ParseMediaType(value)
			if err != nil || !strings.Contains(mediaType, "/") || strings.Contains(mediaType, "*") {
				op.Warnings = append(op.Warnings, "invalid response-content-type \""+value+"\"; expected a media type such as application/octet-stream; ignored")
				continue
			}
			op.ResponseContentType = value
		case "response-format":
			if value != "binary" {
				op.Warnings = append(op.Warnings, "invalid response-format \""+value+"\"; supported: binary; ignored")
				continue
			}
			op.ResponseFormat = value
		case "content-type":
			switch strings.ToLower(value) {
			case "multipart", "multipart/form-data":
				op.ContentType = "multipart/form-data"
			case "form", "urlencoded", "application/x-www-form-urlencoded":
				op.ContentType = "application/x-www-form-urlencoded"
			default:
				op.ContentType = value
			}
		default:

			if looksLikeKey(key) {
				op.Warnings = append(op.Warnings,
					"unknown annotation key \""+key+"\"; ignored (did you mean one of summary, description, tags, errors, security, request, response, content-type, response-status, response-content-type, response-format?)")
			}
		}
	}

	if !found {
		return nil, false
	}
	return op, true
}
