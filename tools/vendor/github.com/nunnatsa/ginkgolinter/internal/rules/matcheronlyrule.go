package rules

// matcherOnlyRules is a collection of rules that only validate the matcher part of the assertion.
// It does not validate the actual part of the assertion.
var matcherOnlyRules = Rules{
	&HaveLen0{},
	&EqualBoolRule{},
	&EqualNilRule{},
	&DoubleNegativeRule{},
}

func getMatcherOnlyRules() Rules {
	return matcherOnlyRules
}
