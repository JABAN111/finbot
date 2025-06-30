package main

const (
	beginOperation = iota
	settingSum
	choosingStatus
	choosingCategory
	choosingComment
)

type UserState struct {
	UserStateCurrentOperation int
	OperationSum              float64
	Category                  string
	Status                    OperationStatus
	Comment                   string
}

func (us *UserState) Reset() {
	us.UserStateCurrentOperation = beginOperation
	us.OperationSum = 0
	us.Category = ""
	us.Status = ""
	us.Comment = ""
}
