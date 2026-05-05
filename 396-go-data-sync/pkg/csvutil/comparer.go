package csvutil

import "strings"

func CompareCSV(source *CSVData, target *CSVData) []Change {
	var changes []Change
	
	if target == nil {
		for key, record := range source.Records {
			newRec := record
			changes = append(changes, Change{
				Type:      ChangeAdd,
				Key:       key,
				NewRecord: &newRec,
			})
		}
		return changes
	}
	
	for key, sourceRec := range source.Records {
		targetRec, exists := target.Records[key]
		if !exists {
			newRec := sourceRec
			changes = append(changes, Change{
				Type:      ChangeAdd,
				Key:       key,
				NewRecord: &newRec,
			})
		} else {
			if !recordsEqual(&sourceRec, &targetRec) {
				oldRec := targetRec
				newRec := sourceRec
				changes = append(changes, Change{
					Type:      ChangeModify,
					Key:       key,
					OldRecord: &oldRec,
					NewRecord: &newRec,
				})
			}
		}
	}
	
	for key, targetRec := range target.Records {
		_, exists := source.Records[key]
		if !exists {
			oldRec := targetRec
			changes = append(changes, Change{
				Type:      ChangeDelete,
				Key:       key,
				OldRecord: &oldRec,
			})
		}
	}
	
	return changes
}

func recordsEqual(r1, r2 *CSVRecord) bool {
	if len(r1.Values) != len(r2.Values) {
		return false
	}
	
	for key, v1 := range r1.Values {
		v2, ok := r2.Values[key]
		if !ok {
			return false
		}
		if strings.TrimSpace(v1) != strings.TrimSpace(v2) {
			return false
		}
	}
	
	return true
}
