package w32uiautomation

func NewTreeWalker(auto *IUIAutomation) (*IUIAutomationTreeWalker, error) {
	condition, err := auto.CreateTrueCondition()
	if err != nil {
		return nil, err
	}
	defer condition.Release()

	return auto.CreateTreeWalker(condition)
}
