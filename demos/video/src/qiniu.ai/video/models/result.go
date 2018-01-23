package models

type ResultBody struct {
	Result []Result `bson:"result" json:"result"`
	Type   string   `bson:"type" json:"type"`
}

func MergeResult(resList []Result) (mergeList []Result) {
	mergeList = make([]Result, 0, len(resList)/2)

	for _, res := range resList {
		flag := true
		for i := len(mergeList) - 1; i >= 0; i-- {
			if res.Attribute == mergeList[i].Attribute {
				if res.Time.Start == mergeList[i].Time.End+1 || res.Time.Start == mergeList[i].Time.End {
					mergeList[i].Confidence = (mergeList[i].Confidence*float64(mergeList[i].Time.End-mergeList[i].Time.Start+1) + res.Confidence) / float64(mergeList[i].Time.End-mergeList[i].Time.Start+2)
					if mergeList[i].Time.End < res.Time.End {
						mergeList[i].Time.End = res.Time.End
					}
					flag = false
					break
				} else if res.Time.Start < mergeList[i].Time.Start-1 && res.Time.Start > mergeList[i].Time.End+1 {
					break
				}

			}

		}
		if flag {
			mergeList = append(mergeList, res)
		}

	}

	return mergeList

}

func FilterScene(srcList []Result) (resultList []Result) {
	if len(srcList) == 0 {
		return srcList
	}
	for _, src := range srcList {
		if src.Time.Start == src.Time.End || src.Time.End == src.Time.Start+1 {
			continue
		}
		resultList = append(resultList, src)
	}
	return
}

type Result struct {
	Attribute  string       `bson:"attribute" json:"attribute"`
	Confidence float64      `bson:"confidence" json:"confidence"`
	Type       string       `bson:"type" json:"type"`
	Time       TimeDuration `bson:"time" json:"time"`
}

type TimeDuration struct {
	Start int `bson:"start" json:"start"`
	End   int `bson:"end" json:"end"`
}

type TaskStatus string

const (
	TaskStatusDone       TaskStatus = "DONE"
	TaskStatusError      TaskStatus = "ERROR"
	TaskstatusPorcessing TaskStatus = "PROCESSING"
)

func (tstatus TaskStatus) IsValid() bool {
	switch tstatus {
	case TaskStatusDone, TaskStatusError, TaskstatusPorcessing:
		return true
	}

	return false
}

func confidenceAbs(a float64, b float64) bool {
	if a-b >= -0.15 && a-b <= 0.15 && a < 0.6 && b < 0.6 {
		return true
	}
	return false
}
