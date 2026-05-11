package samlparser

import "encoding/xml"

const (
	SAMLAssertionNS   = "urn:oasis:names:tc:SAML:2.0:assertion"
	SAMLProtocolNS    = "urn:oasis:names:tc:SAML:2.0:protocol"
	DSigNS            = "http://www.w3.org/2000/09/xmldsig#"
	BearerMethod      = "urn:oasis:names:tc:SAML:2.0:cm:bearer"
	ClockSkewMinutes  = 2
)

type Assertion struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	ID      string   `xml:",attr"`
	Version string   `xml:",attr"`

	Subject  *Subject  `xml:"Subject"`
	Conditions *Conditions `xml:"Conditions"`
	
	AttributeStatements []AttributeStatement `xml:"AttributeStatement"`
	Signatures          []Signature         `xml:"Signature"`
}

type Subject struct {
	NameID             *NameID              `xml:"NameID"`
	SubjectConfirmation *SubjectConfirmation `xml:"SubjectConfirmation"`
}

type NameID struct {
	Value string `xml:",chardata"`
}

type SubjectConfirmation struct {
	Method string                   `xml:",attr"`
	Data   *SubjectConfirmationData `xml:"SubjectConfirmationData"`
}

type SubjectConfirmationData struct {
	NotOnOrAfter string `xml:",attr"`
	Recipient    string `xml:",attr"`
	InResponseTo string `xml:",attr"`
	Address      string `xml:",attr"`
}

type Conditions struct {
	NotBefore            string               `xml:",attr"`
	NotOnOrAfter         string               `xml:",attr"`
	AudienceRestrictions []AudienceRestriction `xml:"AudienceRestriction"`
}

type AudienceRestriction struct {
	Audiences []Audience `xml:"Audience"`
}

type Audience struct {
	Value string `xml:",chardata"`
}

type AttributeStatement struct {
	Attributes []Attribute `xml:"Attribute"`
}

type Attribute struct {
	Name          string            `xml:",attr"`
	FriendlyName  string            `xml:",attr"`
	NameFormat    string            `xml:",attr"`
	AttributeValues []AttributeValue `xml:"AttributeValue"`
}

type AttributeValue struct {
	Type  string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Value string `xml:",chardata"`
}

type Signature struct {
	XMLName        xml.Name       `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
	SignedInfo     *SignedInfo    `xml:"SignedInfo"`
	SignatureValue string         `xml:"SignatureValue"`
	KeyInfo        *KeyInfo       `xml:"KeyInfo"`
}

type SignedInfo struct {
	CanonicalizationMethod *CanonicalizationMethod `xml:"CanonicalizationMethod"`
	SignatureMethod        *SignatureMethod        `xml:"SignatureMethod"`
	References             []Reference             `xml:"Reference"`
}

type CanonicalizationMethod struct {
	Algorithm string `xml:",attr"`
}

type SignatureMethod struct {
	Algorithm string `xml:",attr"`
}

type Reference struct {
	URI          string       `xml:",attr"`
	Transforms   *Transforms  `xml:"Transforms"`
	DigestMethod *DigestMethod `xml:"DigestMethod"`
	DigestValue  string       `xml:"DigestValue"`
}

type Transforms struct {
	Transforms []Transform `xml:"Transform"`
}

type Transform struct {
	Algorithm string `xml:",attr"`
}

type DigestMethod struct {
	Algorithm string `xml:",attr"`
}

type KeyInfo struct {
	X509Data    *X509Data    `xml:"X509Data"`
	KeyValue    *KeyValue    `xml:"KeyValue"`
}

type X509Data struct {
	X509Certificates []X509Certificate `xml:"X509Certificate"`
}

type X509Certificate struct {
	Value string `xml:",chardata"`
}

type KeyValue struct {
	RSAKeyValue *RSAKeyValue `xml:"RSAKeyValue"`
}

type RSAKeyValue struct {
	Modulus  string `xml:"Modulus"`
	Exponent string `xml:"Exponent"`
}
