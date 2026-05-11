package vcard

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"vcard-parser/api"
)

type Property struct {
	Group    string
	Name     string
	Params   map[string][]string
	Value    string
	Version  string
}

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(r),
	}
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && line != "" {
			return strings.TrimRight(line, "\r\n"), nil
		}
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (p *Parser) readAllLines() ([]string, error) {
	var lines []string
	for {
		line, err := p.readLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func unfoldLines(lines []string) []string {
	var unfolded []string
	var current strings.Builder

	for _, line := range lines {
		if line == "" {
			continue
		}
		if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && current.Len() > 0 {
			if len(line) > 0 {
				trimmed := strings.TrimLeft(line, " \t")
				if strings.HasSuffix(current.String(), "=") {
					current.WriteString(trimmed)
				} else {
					current.WriteString(trimmed)
				}
			}
		} else {
			if current.Len() > 0 {
				unfolded = append(unfolded, current.String())
			}
			current.Reset()
			current.WriteString(line)
		}
	}
	if current.Len() > 0 {
		unfolded = append(unfolded, current.String())
	}

	return unfolded
}

func parseProperty(line string, version string) (*Property, error) {
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	group := ""
	name := ""
	params := make(map[string][]string)
	value := ""

	colonIdx := -1
	for i := 0; i < len(line); i++ {
		if i+1 < len(line) && line[i] == '\\' {
			i++
			continue
		}
		if line[i] == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid property: %s", line)
	}

	propPart := line[:colonIdx]
	value = line[colonIdx+1:]

	dotIdx := -1
	for i := 0; i < len(propPart); i++ {
		if i+1 < len(propPart) && propPart[i] == '\\' {
			i++
			continue
		}
		if propPart[i] == '.' {
			dotIdx = i
			break
		}
	}

	if dotIdx != -1 {
		group = propPart[:dotIdx]
		propPart = propPart[dotIdx+1:]
	}

	parts := splitByChar(propPart, ';', true)
	name = parts[0]

	if len(parts) > 1 {
		for _, paramPart := range parts[1:] {
			if paramPart == "" {
				continue
			}
			equalsIdx := strings.Index(paramPart, "=")
			if equalsIdx == -1 {
				if strings.EqualFold(paramPart, "QUOTED-PRINTABLE") {
					params["ENCODING"] = append(params["ENCODING"], "QUOTED-PRINTABLE")
				} else if strings.EqualFold(paramPart, "BASE64") {
					params["ENCODING"] = append(params["ENCODING"], "BASE64")
				} else if strings.EqualFold(paramPart, "B") {
					params["ENCODING"] = append(params["ENCODING"], "b")
				} else {
					params["TYPE"] = append(params["TYPE"], paramPart)
				}
				continue
			}

			paramName := strings.ToUpper(paramPart[:equalsIdx])
			paramValue := paramPart[equalsIdx+1:]

			if strings.HasPrefix(paramValue, "\"") && strings.HasSuffix(paramValue, "\"") {
				paramValue = paramValue[1 : len(paramValue)-1]
			}

			if strings.EqualFold(paramName, "TYPE") {
				values := strings.Split(paramValue, ",")
				for _, v := range values {
					params[paramName] = append(params[paramName], v)
				}
			} else {
				params[paramName] = append(params[paramName], paramValue)
			}
		}
	}

	return &Property{
		Group:   group,
		Name:    strings.ToUpper(name),
		Params:  params,
		Value:   value,
		Version: version,
	}, nil
}

func splitVCardBlocks(lines []string) [][]string {
	var blocks [][]string
	var currentBlock []string
	inBlock := false

	for _, line := range lines {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if upperLine == "BEGIN:VCARD" {
			inBlock = true
			currentBlock = []string{line}
		} else if upperLine == "END:VCARD" {
			if inBlock {
				currentBlock = append(currentBlock, line)
				blocks = append(blocks, currentBlock)
				currentBlock = nil
				inBlock = false
			}
		} else if inBlock {
			currentBlock = append(currentBlock, line)
		}
	}

	return blocks
}

func detectVersion(properties []*Property) string {
	for _, prop := range properties {
		if prop.Name == "VERSION" {
			version := strings.TrimSpace(prop.Value)
			if version == "2.1" || version == "3.0" || version == "4.0" {
				return version
			}
		}
	}
	return "3.0"
}

func parseBlock(block []string) (*api.Contact, error) {
	unfolded := unfoldLines(block)

	var properties []*Property
	for _, line := range unfolded {
		prop, err := parseProperty(line, "3.0")
		if err != nil {
			continue
		}
		properties = append(properties, prop)
	}

	version := detectVersion(properties)

	contact := &api.Contact{
		Version:     version,
		ExtraFields: make(map[string][]string),
	}

	for _, prop := range properties {
		prop.Version = version
		processProperty(contact, prop)
	}

	return contact, nil
}

func processProperty(contact *api.Contact, prop *Property) {
	switch prop.Name {
	case "FN":
		contact.FullName = unescapeValue(prop.Value)
	case "N":
		contact.Name = parseName(prop.Value)
	case "TEL":
		phone := parsePhone(prop)
		contact.Phones = append(contact.Phones, phone)
	case "EMAIL":
		email := parseEmail(prop)
		contact.Emails = append(contact.Emails, email)
	case "ORG":
		parts := splitPropertyValueWithEscaping(prop.Value, ';')
		if len(parts) > 0 {
			contact.Organization = unescapeValue(parts[0])
		}
	case "TITLE":
		contact.Title = unescapeValue(prop.Value)
	case "ADR":
		addr := parseAddress(prop)
		contact.Addresses = append(contact.Addresses, addr)
	case "BDAY":
		contact.Birthday = unescapeValue(prop.Value)
	case "PHOTO", "LOGO":
		photo := parsePhoto(prop)
		if photo != nil {
			contact.Photo = photo
		}
	case "GROUP":
		contact.Group = unescapeValue(prop.Value)
	case "NOTE":
		contact.Note = unescapeValue(prop.Value)
	case "URL":
		contact.URL = unescapeValue(prop.Value)
	case "VERSION", "BEGIN", "END":
	default:
		contact.ExtraFields[prop.Name] = append(contact.ExtraFields[prop.Name], unescapeValue(prop.Value))
	}

	if prop.Group != "" && contact.Group == "" {
		contact.Group = prop.Group
	}
}

func parseName(value string) *api.Name {
	parts := splitPropertyValueWithEscaping(value, ';')
	for len(parts) < 5 {
		parts = append(parts, "")
	}

	additional := []string{}
	if parts[2] != "" {
		for _, a := range strings.Split(unescapeValue(parts[2]), ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				additional = append(additional, a)
			}
		}
	}

	prefixes := []string{}
	if parts[3] != "" {
		for _, p := range strings.Split(unescapeValue(parts[3]), ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				prefixes = append(prefixes, p)
			}
		}
	}

	suffixes := []string{}
	if parts[4] != "" {
		for _, s := range strings.Split(unescapeValue(parts[4]), ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				suffixes = append(suffixes, s)
			}
		}
	}

	return &api.Name{
		Family:     unescapeValue(parts[0]),
		Given:      unescapeValue(parts[1]),
		Additional: additional,
		Prefixes:   prefixes,
		Suffixes:   suffixes,
	}
}

func parsePhone(prop *Property) api.Phone {
	types := []string{}
	if typeValues, ok := prop.Params["TYPE"]; ok {
		for _, t := range typeValues {
			types = append(types, strings.ToUpper(t))
		}
	}
	return api.Phone{
		Number: unescapeValue(prop.Value),
		Types:  types,
	}
}

func parseEmail(prop *Property) api.Email {
	types := []string{}
	if typeValues, ok := prop.Params["TYPE"]; ok {
		for _, t := range typeValues {
			types = append(types, strings.ToUpper(t))
		}
	}
	return api.Email{
		Address: unescapeValue(prop.Value),
		Types:   types,
	}
}

func parseAddress(prop *Property) api.Address {
	types := []string{}
	if typeValues, ok := prop.Params["TYPE"]; ok {
		for _, t := range typeValues {
			types = append(types, strings.ToUpper(t))
		}
	}

	label := ""
	if labelValues, ok := prop.Params["LABEL"]; ok && len(labelValues) > 0 {
		label = unescapeValue(labelValues[0])
	}

	parts := splitPropertyValueWithEscaping(prop.Value, ';')
	for len(parts) < 7 {
		parts = append(parts, "")
	}

	return api.Address{
		Type:          types,
		PostOfficeBox: unescapeValue(parts[0]),
		Extended:      unescapeValue(parts[1]),
		Street:        unescapeValue(parts[2]),
		Locality:      unescapeValue(parts[3]),
		Region:        unescapeValue(parts[4]),
		PostalCode:    unescapeValue(parts[5]),
		Country:       unescapeValue(parts[6]),
		Label:         label,
	}
}

func parsePhoto(prop *Property) *api.Photo {
	photo := &api.Photo{}

	if encodingValues, ok := prop.Params["ENCODING"]; ok && len(encodingValues) > 0 {
		photo.Encoding = strings.ToUpper(encodingValues[0])
	}

	if typeValues, ok := prop.Params["TYPE"]; ok && len(typeValues) > 0 {
		mediaType := typeValues[0]
		if !strings.Contains(mediaType, "/") && !strings.HasPrefix(mediaType, "image/") {
			mediaType = "image/" + mediaType
		}
		photo.MediaType = mediaType
	}

	if mediaTypeValues, ok := prop.Params["MEDIATYPE"]; ok && len(mediaTypeValues) > 0 {
		photo.MediaType = mediaTypeValues[0]
	}

	if valueValues, ok := prop.Params["VALUE"]; ok && len(valueValues) > 0 {
		valueType := strings.ToUpper(valueValues[0])
		if valueType == "URI" || valueType == "URL" {
			photo.URL = prop.Value
			return photo
		}
	}

	if strings.HasPrefix(strings.ToLower(prop.Value), "http://") || strings.HasPrefix(strings.ToLower(prop.Value), "https://") {
		photo.URL = prop.Value
		return photo
	}

	cleanValue := prop.Value
	cleanValue = strings.ReplaceAll(cleanValue, "\n", "")
	cleanValue = strings.ReplaceAll(cleanValue, "\r", "")
	cleanValue = strings.ReplaceAll(cleanValue, " ", "")
	cleanValue = strings.ReplaceAll(cleanValue, "\t", "")

	if photo.Encoding == "BASE64" || photo.Encoding == "B" || isValidBase64(cleanValue) {
		photo.Base64Data = cleanValue
		return photo
	}

	if prop.Value != "" {
		photo.URL = prop.Value
	}

	return photo
}

func isValidBase64(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if len(s) < 4 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func (p *Parser) Parse() ([]api.Contact, error) {
	lines, err := p.readAllLines()
	if err != nil {
		return nil, fmt.Errorf("failed to read lines: %v", err)
	}

	blocks := splitVCardBlocks(lines)

	if len(blocks) == 0 {
		return []api.Contact{}, nil
	}

	contacts := []api.Contact{}
	for _, block := range blocks {
		contact, err := parseBlock(block)
		if err != nil {
			continue
		}
		contacts = append(contacts, *contact)
	}

	return contacts, nil
}

func ParseString(vcardText string) ([]api.Contact, error) {
	reader := strings.NewReader(vcardText)
	parser := NewParser(reader)
	return parser.Parse()
}
