package samlparser

import (
	"bytes"
	"encoding/xml"
	"strings"
	"time"

	"saml-parser/pkg/model"
)

func Parse(xmlContent string) (*model.ParseResponse, error) {
	var assertion Assertion
	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlContent)))
	
	if err := decoder.Decode(&assertion); err != nil {
		return &model.ParseResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	response := &model.ParseResponse{
		Success:    true,
		Attributes: extractAttributes(&assertion),
		Signature:  extractSignature(&assertion),
		Subject:    extractSubject(&assertion),
		Conditions: extractConditions(&assertion),
	}

	return response, nil
}

func ExtractAttribute(xmlContent string, attributeName string) (*model.ExtractResponse, error) {
	parseResp, err := Parse(xmlContent)
	if err != nil {
		return &model.ExtractResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	for _, attr := range parseResp.Attributes {
		if strings.EqualFold(attr.Name, attributeName) {
			return &model.ExtractResponse{
				Success: true,
				Values:  attr.Values,
			}, nil
		}
	}

	return &model.ExtractResponse{
		Success: true,
		Values:  []string{},
	}, nil
}

func extractAttributes(assertion *Assertion) []model.Attribute {
	result := []model.Attribute{}

	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			values := []string{}
			for _, val := range attr.AttributeValues {
				trimmed := strings.TrimSpace(val.Value)
				if trimmed != "" {
					values = append(values, trimmed)
				}
			}
			
			if len(values) > 0 {
				result = append(result, model.Attribute{
					Name:   attr.Name,
					Values: values,
				})
			}
		}
	}

	return result
}

func extractSignature(assertion *Assertion) model.SignatureInfo {
	sigInfo := model.SignatureInfo{
		HasSignature: false,
	}

	if len(assertion.Signatures) > 0 {
		sig := assertion.Signatures[0]
		sigInfo.HasSignature = true
		
		if sig.SignedInfo != nil && sig.SignedInfo.SignatureMethod != nil {
			sigInfo.Algorithm = sig.SignedInfo.SignatureMethod.Algorithm
		}
		
		sigInfo.SignatureValue = sig.SignatureValue
		
		if sig.SignedInfo != nil && len(sig.SignedInfo.References) > 0 {
			sigInfo.DigestValue = sig.SignedInfo.References[0].DigestValue
		}
	}

	return sigInfo
}

func extractSubject(assertion *Assertion) model.SubjectInfo {
	subjectInfo := model.SubjectInfo{}

	if assertion.Subject != nil {
		if assertion.Subject.NameID != nil {
			subjectInfo.NameID = assertion.Subject.NameID.Value
		}
		
		if assertion.Subject.SubjectConfirmation != nil {
			sc := assertion.Subject.SubjectConfirmation
			subjectInfo.SubjectConfirmation = model.SubjectConfirmationInfo{
				Method: sc.Method,
			}
			
			if sc.Data != nil {
				subjectInfo.SubjectConfirmation.Data = model.SubjectConfirmationDataInfo{
					Recipient:    sc.Data.Recipient,
					NotOnOrAfter: sc.Data.NotOnOrAfter,
					InResponseTo: sc.Data.InResponseTo,
				}
			}
		}
	}

	return subjectInfo
}

func extractConditions(assertion *Assertion) model.ConditionsInfo {
	conditionsInfo := model.ConditionsInfo{}

	if assertion.Conditions != nil {
		conditionsInfo.NotBefore = assertion.Conditions.NotBefore
		conditionsInfo.NotOnOrAfter = assertion.Conditions.NotOnOrAfter
		
		for _, ar := range assertion.Conditions.AudienceRestrictions {
			for _, audience := range ar.Audiences {
				conditionsInfo.AudienceRestriction.Audiences = append(
					conditionsInfo.AudienceRestriction.Audiences,
					audience.Value,
				)
			}
		}
	}

	return conditionsInfo
}

func parseSAMLTime(timeStr string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999999Z",
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, nil
}
