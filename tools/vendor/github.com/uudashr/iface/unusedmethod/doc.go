// Package unusedmethod defines an Analyzer that detects interface methods which
// are never used anywhere in the same package where they are defined.
//
// A method is considered used only when it is invoked or referenced through a
// value of the interface type. Merely implementing the interface — for example,
// assigning a concrete type to the interface, or calling the method directly on
// the concrete type — does not count as a use.
package unusedmethod
