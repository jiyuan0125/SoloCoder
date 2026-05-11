package api

type Name struct {
	Family     string   `json:"family,omitempty"`
	Given      string   `json:"given,omitempty"`
	Additional []string `json:"additional,omitempty"`
	Prefixes   []string `json:"prefixes,omitempty"`
	Suffixes   []string `json:"suffixes,omitempty"`
}

type Phone struct {
	Number string   `json:"number"`
	Types  []string `json:"types,omitempty"`
}

type Email struct {
	Address string   `json:"address"`
	Types   []string `json:"types,omitempty"`
}

type Address struct {
	Type        []string `json:"type,omitempty"`
	PostOfficeBox string   `json:"postOfficeBox,omitempty"`
	Extended    string   `json:"extended,omitempty"`
	Street      string   `json:"street,omitempty"`
	Locality    string   `json:"locality,omitempty"`
	Region      string   `json:"region,omitempty"`
	PostalCode  string   `json:"postalCode,omitempty"`
	Country     string   `json:"country,omitempty"`
	Label       string   `json:"label,omitempty"`
}

type Photo struct {
	URL        string `json:"url,omitempty"`
	Base64Data string `json:"base64Data,omitempty"`
	MediaType  string `json:"mediaType,omitempty"`
	Encoding   string `json:"encoding,omitempty"`
}

type Contact struct {
	Version        string    `json:"version"`
	FullName       string    `json:"fullName,omitempty"`
	Name           *Name     `json:"name,omitempty"`
	Phones         []Phone   `json:"phones,omitempty"`
	Emails         []Email   `json:"emails,omitempty"`
	Organization   string    `json:"organization,omitempty"`
	Title          string    `json:"title,omitempty"`
	Addresses      []Address `json:"addresses,omitempty"`
	Birthday       string    `json:"birthday,omitempty"`
	Photo          *Photo    `json:"photo,omitempty"`
	Group          string    `json:"group,omitempty"`
	Note           string    `json:"note,omitempty"`
	URL            string    `json:"url,omitempty"`
	ExtraFields    map[string][]string `json:"extraFields,omitempty"`
}
