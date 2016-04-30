package metric

import (
	"fmt"

	"edgeg.io/gtm/event"
	"edgeg.io/gtm/note"
	"edgeg.io/gtm/project"
	"edgeg.io/gtm/report"
	"edgeg.io/gtm/scm"
)

type GitState int

const (
	Working GitState = iota
	Staging
	Committed
)

func Process(gstate GitState, debug bool) (string, error) {

	_, gtmPath, err := project.Paths()
	if err != nil {
		return "", err
	}

	// load any saved metrics
	metricMap, err := loadMetrics(gtmPath)
	if err != nil {
		return "", err
	}

	// process event files
	epochEventMap, err := event.Process(gtmPath, gstate == Working || gstate == Staging)
	if err != nil {
		return "", err
	}

	// allocate time for events
	for ep := range epochEventMap {
		err := allocateTime(ep, metricMap, epochEventMap[ep])
		if err != nil {
			return "", err
		}
	}

	// build map of commit files
	commitMap, err := buildCommitMap(metricMap, gstate)
	if err != nil {
		return "", err
	}

	// create time logged struct
	logged, err := buildCommitNote(metricMap, commitMap, gstate)
	if err != nil {
		return "", err
	}

	msg := ""
	if gstate == Working || gstate == Staging {
		msg, err = report.NoteFiles(logged)
		if err != nil {
			return "", err
		}
	}
	if gstate == Committed {
		if err := scm.GitAddNote(note.Marshal(logged), project.NoteNameSpace); err != nil {
			return "", err
		}
		if err := saveAndPurgeMetrics(gtmPath, metricMap, commitMap); err != nil {
			return "", err
		}
	}

	if debug {
		msg += fmt.Sprintf("\nEventMap:\n%+v\n", epochEventMap)
		msg += fmt.Sprintf("\nMetricMap:\n%+v\n", metricMap)
		msg += fmt.Sprintf("\nCommitMap:\n%+v\n", commitMap)
	}

	return msg, nil
}

func buildCommitMap(metricMap map[string]FileMetric, gstate GitState) (map[string]FileMetric, error) {
	commitMap := map[string]FileMetric{}

	if gstate == Committed {
		m, err := scm.GitLastLog()
		if err != nil {
			return commitMap, err
		}
		_, _, commitFiles := scm.GitParseMessage(m)
		for _, f := range commitFiles {
			fileID := getFileID(f)
			if _, ok := metricMap[fileID]; !ok {
				continue
			}
			commitMap[fileID] = metricMap[fileID]
		}
	} else {
		// include git tracked files that have been modified
		for fileID, fm := range metricMap {
			if fm.GitTracked {
				modified, err := scm.GitModified(fm.SourceFile, gstate == Staging)
				if err != nil {
					return commitMap, err
				}
				if modified {
					commitMap[fileID] = fm
				}
			}
		}
	}

	return commitMap, nil
}