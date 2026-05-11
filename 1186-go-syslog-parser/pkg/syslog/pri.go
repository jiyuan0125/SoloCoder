package syslog

import (
	"fmt"
	"strconv"
)

type Facility int

const (
	FacilityKern     Facility = 0
	FacilityUser     Facility = 1
	FacilityMail     Facility = 2
	FacilityDaemon   Facility = 3
	FacilityAuth     Facility = 4
	FacilitySyslog   Facility = 5
	FacilityLpr      Facility = 6
	FacilityNews     Facility = 7
	FacilityUUCP     Facility = 8
	FacilityCron     Facility = 9
	FacilityAuthPriv Facility = 10
	FacilityFTP      Facility = 11
	FacilityNTP      Facility = 12
	FacilityLogAudit Facility = 13
	FacilityLogAlert Facility = 14
	FacilityClock    Facility = 15
	FacilityLocal0   Facility = 16
	FacilityLocal1   Facility = 17
	FacilityLocal2   Facility = 18
	FacilityLocal3   Facility = 19
	FacilityLocal4   Facility = 20
	FacilityLocal5   Facility = 21
	FacilityLocal6   Facility = 22
	FacilityLocal7   Facility = 23
)

type Severity int

const (
	SeverityEmergency Severity = 0
	SeverityAlert     Severity = 1
	SeverityCritical  Severity = 2
	SeverityError     Severity = 3
	SeverityWarning   Severity = 4
	SeverityNotice    Severity = 5
	SeverityInfo      Severity = 6
	SeverityDebug     Severity = 7
)

var FacilityNames = map[Facility]string{
	FacilityKern:     "kern",
	FacilityUser:     "user",
	FacilityMail:     "mail",
	FacilityDaemon:   "daemon",
	FacilityAuth:     "auth",
	FacilitySyslog:   "syslog",
	FacilityLpr:      "lpr",
	FacilityNews:     "news",
	FacilityUUCP:     "uucp",
	FacilityCron:     "cron",
	FacilityAuthPriv: "authpriv",
	FacilityFTP:      "ftp",
	FacilityNTP:      "ntp",
	FacilityLogAudit: "logaudit",
	FacilityLogAlert: "logalert",
	FacilityClock:    "clock",
	FacilityLocal0:   "local0",
	FacilityLocal1:   "local1",
	FacilityLocal2:   "local2",
	FacilityLocal3:   "local3",
	FacilityLocal4:   "local4",
	FacilityLocal5:   "local5",
	FacilityLocal6:   "local6",
	FacilityLocal7:   "local7",
}

var SeverityNames = map[Severity]string{
	SeverityEmergency: "emergency",
	SeverityAlert:     "alert",
	SeverityCritical:  "critical",
	SeverityError:     "error",
	SeverityWarning:   "warning",
	SeverityNotice:    "notice",
	SeverityInfo:      "info",
	SeverityDebug:     "debug",
}

func CalculatePRI(facility Facility, severity Severity) int {
	return int(facility)*8 + int(severity)
}

func ParsePRI(pri int) (Facility, Severity, error) {
	if pri < 0 || pri > 191 {
		return 0, 0, fmt.Errorf("invalid PRI value: %d, must be between 0 and 191", pri)
	}
	facility := Facility(pri / 8)
	severity := Severity(pri % 8)
	return facility, severity, nil
}

func ParsePRIString(priStr string) (Facility, Severity, int, error) {
	if len(priStr) < 3 || priStr[0] != '<' {
		return 0, 0, 0, fmt.Errorf("invalid PRI format: %s", priStr)
	}

	endIdx := -1
	for i := 1; i < len(priStr); i++ {
		if priStr[i] == '>' {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return 0, 0, 0, fmt.Errorf("invalid PRI format: missing closing >")
	}

	priNumStr := priStr[1:endIdx]
	pri, err := strconv.Atoi(priNumStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid PRI number: %s", priNumStr)
	}

	facility, severity, err := ParsePRI(pri)
	if err != nil {
		return 0, 0, 0, err
	}

	return facility, severity, pri, nil
}

func FormatPRI(pri int) string {
	return fmt.Sprintf("<%d>", pri)
}
