package rulesengine

type (
	// Operator type is the operators enums type
	Operator string
)

const (
	// Logical
	And    Operator = "AND"
	Or     Operator = "OR"
	Not    Operator = "NOT"
	IfThen Operator = "IF_THEN"

	// Equality
	Eq  Operator = "EQ"
	Neq Operator = "NEQ"

	// Numeric Comparison
	Gt      Operator = "GT"
	Gte     Operator = "GTE"
	Lt      Operator = "LT"
	Lte     Operator = "LTE"
	Between Operator = "BETWEEN"
	In      Operator = "IN"
	NotIn   Operator = "NOT_IN"
	AnyIn   Operator = "ANY_IN"

	// String
	Contains    Operator = "CONTAINS"
	NotContains Operator = "NOT_CONTAINS"
	StartsWith  Operator = "STARTS_WITH"
	EndsWith    Operator = "ENDS_WITH"
	Matches     Operator = "MATCHES"
	LengthEq    Operator = "LENGTH_EQ"
	LengthGt    Operator = "LENGTH_GT"
	LengthLt    Operator = "LENGTH_LT"

	// Boolean
	IsTrue  Operator = "IS_TRUE"
	IsFalse Operator = "IS_FALSE"

	// Date / Time
	Before      Operator = "BEFORE"
	After       Operator = "AFTER"
	DateBetween Operator = "DATE_BETWEEN"
	WithinLast  Operator = "WITHIN_LAST"
	WithinNext  Operator = "WITHIN_NEXT"

	// Arrays
	Any  Operator = "ANY"
	All  Operator = "ALL"
	None Operator = "NONE"

	// Existence / Null
	Exists    Operator = "EXISTS"
	NotExists Operator = "NOT_EXISTS"
	IsNull    Operator = "IS_NULL"
	IsNotNull Operator = "IS_NOT_NULL"

	// Type
	IsNumber Operator = "IS_NUMBER"
	IsString Operator = "IS_STRING"
	IsBool   Operator = "IS_BOOL"
	IsDate   Operator = "IS_DATE"
	IsList   Operator = "IS_LIST"
	IsObject Operator = "IS_OBJECT"

	// Optional: Custom/Script
	Custom Operator = "CUSTOM_FUNC"
	Script Operator = "SCRIPT"
)
